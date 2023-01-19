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
	"math/rand"
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
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var (
	ctx  = context.Background()
	Port = flag.Int("port", constants.COMPUTE_GRPC_SERVER_PORT, "The server port")
)

type workerConfig struct {
	rpsUpper            int // Using int for now, should be float64, can be less than 1
	rpsLower            int // Using int for now, should be float64, can be less than 1
	concurrencyUpper    int
	concurrencyLower    int
	concurrentWorkflows int
	image               string
	logLevel            string
	mode                string
	temporalAddress     string
}

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

	workerConfig := getConfigFromEnv()

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get in cluster config!: %v\n", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create kube client!: %v\n", err.Error())
	}
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

	// Watch nodes
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*5)
	nodeInformer := informerFactory.Core().V1().Nodes().Informer()
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			err := createWorkerPod(obj.(*corev1.Node).Name, workerConfig, clientset)
			if err != nil {
				log.Println("Failed to create worker pod: "+constants.WORKER_POD_PREFIX+obj.(*corev1.Node).Name, workerConfig, err.Error())
			}
		},
		DeleteFunc: func(obj interface{}) {
			err := deleteWorkerPod(obj.(*corev1.Node).Name, clientset)
			if err != nil {
				log.Println("Failed to delete worker pod: "+constants.WORKER_POD_PREFIX+obj.(*corev1.Node).Name, err.Error())
			}
		},
	})

	stop := make(chan struct{})
	defer close(stop)
	informerFactory.Start(stop)

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

func getEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Println("Temporal environment variable not set, using default address.")
		return "", errors.New("ENV variable not set " + key)
	}
	return val, nil
}

func createWorkerPod(hostname string, config workerConfig, clientset *kubernetes.Clientset) error {

	var rpsInt, concurrencyInt int
	var rps, concurrency string

	rand.Seed(time.Now().UnixNano())
	if config.rpsUpper == config.rpsLower {
		rpsInt = config.rpsUpper
	} else {
		rpsInt = rand.Intn((config.rpsUpper - config.rpsLower + 1) + config.rpsLower)
	}
	if config.concurrencyUpper == config.concurrencyLower {
		concurrencyInt = config.concurrencyUpper
	} else {
		concurrencyInt = rand.Intn((config.concurrencyUpper - config.concurrencyLower + 1) + config.concurrencyLower)
	}

	if rpsInt == 0 {
		rpsInt = 1
	}
	if concurrencyInt == 0 {
		concurrencyInt = 1
	}
	rps = strconv.Itoa(rpsInt)
	concurrency = strconv.Itoa(concurrencyInt)
	concurrentWorkflows := strconv.Itoa(config.concurrentWorkflows)
	log.Println("Creating worker for node: " + hostname)
	worker := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.WORKER_POD_PREFIX + hostname,
		},
		Spec: corev1.PodSpec{
			NodeSelector: map[string]string{
				constants.KUBE_NODE_LABEL_KEY: constants.KUBE_NODE_LABEL_VAL,
			},
			Containers: []corev1.Container{
				{
					Name:            constants.WORKER_POD_PREFIX + hostname,
					Image:           config.image,
					ImagePullPolicy: constants.POD_PULL_POLICY_ALWAYS,
					Ports: []corev1.ContainerPort{
						{ContainerPort: constants.PROMETHEUS_PORT},
					},
					Env: []corev1.EnvVar{
						{
							Name: constants.MODE_ENV, Value: config.mode,
						},
						{
							Name: constants.TEMPORAL_ENV, Value: config.temporalAddress,
						},
						{
							Name: constants.TEMPORAL_RPS_ENV, Value: rps,
						},
						{
							Name: constants.TEMPORAL_CONCURRENCY_ENV, Value: concurrency,
						},
						{
							Name: constants.TEMPORAL_CONCURRENT_WORKFLOWS_ENV, Value: concurrentWorkflows,
						},
						{
							Name: constants.LOG_LEVEL_ENV, Value: config.logLevel,
						},
						{
							Name: constants.TEMPORAL_TQ_ENV, Value: hostname,
						},
					},
				},
			},
		},
	}
	_, err := clientset.CoreV1().Pods(constants.TEMPORAL_NAMESPACE).Create(context.Background(), worker, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func deleteWorkerPod(hostname string, clientset *kubernetes.Clientset) error {
	log.Println("Creating worker for node: " + hostname)
	err := clientset.CoreV1().Pods(constants.TEMPORAL_NAMESPACE).Delete(context.Background(), constants.WORKER_POD_PREFIX+hostname, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func getConfigFromEnv() workerConfig {
	temporalAddress, err := getEnv(constants.TEMPORAL_ENV)
	if err != nil {
		temporalAddress = constants.TEMPORAL_ADDRESS
	}
	log.Println("WORKER: Using Temporal address from ENV " + temporalAddress)

	rpsUpper, err := getEnv(constants.TEMPORAL_RPS_UPPER_ENV)
	if err != nil {
		rpsUpper = constants.WORKER_DEFAULT_RPS
	}
	log.Println("WORKER: Using RPS_UPPER from ENV " + rpsUpper)

	rpsLower, err := getEnv(constants.TEMPORAL_RPS_LOWER_ENV)
	if err != nil {
		rpsUpper = constants.WORKER_DEFAULT_RPS
	}
	log.Println("WORKER: Using RPS_LOWER from ENV " + rpsLower)

	concurrencyUpper, err := getEnv(constants.TEMPORAL_CONCURRENCY_UPPER_ENV)
	if err != nil {
		concurrencyUpper = constants.WORKER_DEFAULT_CONCURRENCY
	}
	log.Println("WORKER: Using CONCURRENCY_UPPER from ENV " + concurrencyUpper)

	concurrencyLower, err := getEnv(constants.TEMPORAL_CONCURRENCY_LOWER_ENV)
	if err != nil {
		concurrencyLower = constants.WORKER_DEFAULT_CONCURRENCY
	}
	log.Println("WORKER: Using CONCURRENCY_UPPER from ENV " + concurrencyLower)

	logLevel, err := getEnv(constants.LOG_LEVEL_ENV)
	if err != nil {
		logLevel = constants.LOG_LEVEL_DEFAULT
	}
	log.Println("WORKER: Using log level from ENV " + logLevel)

	mode, err := getEnv(constants.MODE_ENV)
	if err != nil {
		// Default to Alcor mode
		mode = constants.MODE_ALCOR
	}
	log.Println("WORKER: Using mode level from ENV " + mode)

	image, err := getEnv(constants.WORKER_IMAGE_ENV)
	if err != nil {
		image = constants.WORKER_DEFAULT_IMAGE
	}
	log.Println("WORKER: Using Image from ENV " + image)

	concurrentWorkflows, err := getEnv(constants.TEMPORAL_CONCURRENT_WORKFLOWS_ENV)
	if err != nil {
		concurrentWorkflows = constants.WORKER_DEFAULT_CONCURRENT_WORKFLOWS
	}
	log.Println("WORKER: Using MaxConcurrentWorkflows from ENV " + concurrentWorkflows)
	concurrencyLowerInt, err := strconv.Atoi(concurrencyLower)
	if err != nil {
		log.Fatalln("ERROR: Unable to convert concurrencyLower to int", err)
	}
	concurrencyUpperInt, err := strconv.Atoi(concurrencyUpper)
	if err != nil {
		log.Fatalln("ERROR: Unable to convert concurrencyUpper to int", err)
	}
	rpsLowerInt, err := strconv.Atoi(rpsLower)
	if err != nil {
		log.Fatalln("ERROR: Unable to convert rpsLower to int", err)
	}
	rpsUpperInt, err := strconv.Atoi(rpsUpper)
	if err != nil {
		log.Fatalln("ERROR: Unable to convert rpsUpper to int", err)
	}
	concurrentWorkflowsInt, err := strconv.Atoi(concurrentWorkflows)
	if err != nil {
		log.Fatalln("ERROR: Unable to convert concurrentWorkflows to int", err)
	}
	return workerConfig{
		mode:                mode,
		temporalAddress:     temporalAddress,
		rpsUpper:            rpsUpperInt,
		rpsLower:            rpsLowerInt,
		concurrencyUpper:    concurrencyUpperInt,
		concurrencyLower:    concurrencyLowerInt,
		concurrentWorkflows: concurrentWorkflowsInt,
		logLevel:            logLevel,
		image:               image,
	}
}
