/*
MIT License
Copyright(c) 2022 Futurewei Cloud

	Permission is hereby granted,
	free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
	including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
	to whom the Software is furnished to do so, subject to the following conditions:
	The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
	WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/handler"
	"github.com/go-redis/redis/v9"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	ctx  = context.Background()
	Port = flag.Int("port", constants.COMPUTE_GRPC_SERVER_PORT, "The server port")
)

func main() {
	// Connect to temporal
	temporal_address, ok := os.LookupEnv(constants.TEMPORAL_ENV)
	if !ok {
		log.Println("Temporal environment variable not set, using default address.")
		temporal_address = constants.LOCALHOST
	}
	var sb strings.Builder
	sb.WriteString(temporal_address)
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(constants.TEMPORAL_PORT))
	var err error
	log.Printf("Connecting to Temporal server at %s", sb.String())

	namespaceClient, err := client.NewNamespaceClient(client.Options{HostPort: sb.String()})
	if err != nil {
		log.Fatalln("ERROR: Unable to create Temporal client for namespace creation", err)
	}
	_, err = namespaceClient.Describe(ctx, constants.TEMPORAL_NAMESPACE)
	if err != nil {
		log.Println("Temporal namespace " + constants.TEMPORAL_NAMESPACE + " doesn't exist! Creating...")
		retention := time.Duration(time.Hour * 48)
		err = namespaceClient.Register(ctx, &workflowservice.RegisterNamespaceRequest{
			Namespace:                        constants.TEMPORAL_NAMESPACE,
			WorkflowExecutionRetentionPeriod: &retention,
		})
		if err != nil {
			log.Fatalln("ERROR: Unable to create Temporal namespace "+constants.TEMPORAL_NAMESPACE, err)
		}
		namespaceClient.Close()
	}

	log.Println("Successfully created created temporal namespace " + constants.TEMPORAL_NAMESPACE)

	handler.TemporalClient, err = client.Dial(client.Options{
		HostPort:  sb.String(),
		Namespace: constants.TEMPORAL_NAMESPACE,
	})
	if err != nil {
		log.Fatalln("ERROR: Unable to create Temporal client", err)
	}
	log.Println("Successfully connected to Temporal on namespace " + constants.TEMPORAL_NAMESPACE)
	defer handler.TemporalClient.Close()

	//Connect to Redis
	var redisAddress strings.Builder
	redisAddress.WriteString(constants.COMPUTE_REDIS_ADDRESS)
	redisAddress.WriteString(":")
	redisAddress.WriteString(strconv.Itoa(constants.COMPUTE_REDIS_PORT))

	handler.RedisClient = *redis.NewClient(&redis.Options{
		Addr:        redisAddress.String(),
		Password:    "", // no password set
		DB:          0,  // use default DB
		PoolSize:    common.COMPUTE_REDIS_POOL_SIZE,
		PoolTimeout: common.COMPUTE_REDIS_POOL_TIMEOUT,
	})

	err = handler.RedisClient.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		log.Fatalln("ERROR: Unable to create Redis client", err)
	}
	log.Println("Successfully connected to Redis!")
	defer handler.RedisClient.Close()
	createWorkers()
	//Start gRPC Server
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *Port))
	if err != nil {
		log.Fatalln("ERROR: Failed to listen", err)
	}
	gRPCServer := grpc.NewServer(
		grpc.MaxSendMsgSize(constants.GRPC_MAX_SEND_MSG_SIZE),
		grpc.MaxRecvMsgSize(constants.GRPC_MAX_RECV_MSG_SIZE))
	pb.RegisterMerakComputeServiceServer(gRPCServer, &handler.Server{})
	log.Printf("Starting gRPC server. Listening at %v", lis.Addr())
	if err := gRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func createWorkers() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get in cluster config!: %v\n", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create kube client!: %v\n", err.Error())
	}
	// Get all nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalln("Failed to get nodes in the cluster!")
	}
	log.Println("Creating workers for nodes: ", nodes.Items)

	// Label the node that this pod is running on
	node, err := clientset.CoreV1().Nodes().Get(context.Background(), os.Getenv("NODE_NAME"), metav1.GetOptions{})
	if err != nil {
		log.Fatalln("Failed to get node!")
	}
	log.Println("Found this host: ", node.Name)
	labelPatch := fmt.Sprintf(`[{"op":"add","path":"/metadata/labels/%s","value":"%s" }]`, constants.KUBE_NODE_LABEL_KEY, constants.KUBE_NODE_LABEL_VAL)
	_, err = clientset.CoreV1().Nodes().Patch(context.Background(), node.Name, types.JSONPatchType, []byte(labelPatch), metav1.PatchOptions{})
	if err != nil {
		log.Fatalln("Failed to label node!")
	}

	log.Println("Labeled this host: ", node.Name)

	// Get ENV variables
	temporal_addr, err := getEnv(constants.TEMPORAL_ENV)
	if err != nil {
		temporal_addr = constants.TEMPORAL_ADDRESS
	}
	log.Println("WORKER: Using Temporal address from ENV " + constants.TEMPORAL_ADDRESS)

	rps, err := getEnv(constants.TEMPORAL_RPS_ENV)
	if err != nil {
		rps = constants.WORKER_DEFAULT_RPS
	}
	log.Println("WORKER: Using RPS from ENV " + rps)

	concurrency, err := getEnv(constants.TEMPORAL_CONCURRENCY_ENV)
	if err != nil {
		concurrency = constants.WORKER_DEFAULT_CONCURRENCY
	}
	log.Println("WORKER: Using CONCURRENCY from ENV " + concurrency)

	logLevel, err := getEnv(constants.LOG_LEVEL_ENV)
	if err != nil {

		logLevel = constants.LOG_LEVEL_DEFAULT
	}
	log.Println("WORKER: Using log level from ENV " + logLevel)

	mode, err := getEnv(constants.MODE_ENV)
	if err != nil {
		mode = constants.MODE_ALCOR
	}
	log.Println("WORKER: Using log level from ENV " + logLevel)

	image, err := getEnv(constants.WORKER_IMAGE_ENV)
	if err != nil {
		log.Println("Image ENV not set! Using default mode " + constants.WORKER_DEFAULT_IMAGE)
		image = constants.WORKER_DEFAULT_IMAGE
	}
	log.Println("WORKER: Using image from ENV " + image)

	// Create worker pods, one per node (only nodes hosting vhosts)
	for _, host := range nodes.Items {
		worker := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.WORKER_POD_PREFIX + host.Name,
			},
			Spec: corev1.PodSpec{
				NodeSelector: map[string]string{
					constants.KUBE_NODE_LABEL_KEY: constants.KUBE_NODE_LABEL_VAL,
				},
				Containers: []corev1.Container{
					{
						Name:            constants.WORKER_POD_PREFIX + host.Name,
						Image:           image,
						ImagePullPolicy: constants.POD_PULL_POLICY_ALWAYS,
						Ports: []corev1.ContainerPort{
							{ContainerPort: constants.PROMETHEUS_PORT},
						},
						Env: []corev1.EnvVar{
							{
								Name: constants.MODE_ENV, Value: mode,
							},
							{
								Name: constants.TEMPORAL_ENV, Value: temporal_addr,
							},
							{
								Name: constants.TEMPORAL_RPS_ENV, Value: rps,
							},
							{
								Name: constants.TEMPORAL_CONCURRENCY_ENV, Value: concurrency,
							},
							{
								Name: constants.LOG_LEVEL_ENV, Value: logLevel,
							},
							{
								Name: constants.TEMPORAL_TQ_ENV, Value: host.Name,
							},
						},
					},
				},
			},
		}
		_, err := clientset.CoreV1().Pods(constants.TEMPORAL_NAMESPACE).Create(context.Background(), worker, metav1.CreateOptions{})
		if err != nil {
			log.Fatalln("Failed to create worker pod: "+constants.WORKER_POD_PREFIX+host.Name, err.Error())
		}
	}

}

func getEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Println("Temporal environment variable not set, using default address.")
		return "", errors.New("ENV variable not set " + key)
	}
	return val, nil
}
