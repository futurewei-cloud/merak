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
package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/go-redis/redis/v9"
	"go.temporal.io/sdk/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	workflowOptions client.StartWorkflowOptions
	TemporalClient  client.Client
	RedisClient     redis.Client
)

type Server struct {
	pb.UnimplementedMerakComputeServiceServer
}

func (s *Server) ComputeHandler(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnComputeMessage, error) {
	log.Println("Received on ComputeHandler", in)
	createWorkers(in)
	switch op := in.OperationType; op {
	case common_pb.OperationType_INFO:

		return caseInfo(ctx, in)

	case common_pb.OperationType_CREATE:

		return caseCreate(ctx, in)

	case common_pb.OperationType_DELETE:

		return caseDelete(ctx, in)

	default:
		log.Println("Unknown Operation")
		return &pb.ReturnComputeMessage{
			ReturnMessage: "Unknown Operation",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, errors.New("unknown operation")
	}
}

// Get all nodes that the pods are scheduled on
// Create a task queue for each node
func createWorkers(in *pb.InternalComputeConfigInfo) {
	log.Println("Creating workers for nodes: ", in.Config.VmDeploy.Hosts)

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
	labelPatch := fmt.Sprintf(`[{"op":"add","path":"/metadata/labels/%s","value":"%s" }]`, constants.KUBE_NODE_LABEL_KEY, constants.KUBE_NODE_LABEL_VAL)
	_, err = clientset.CoreV1().Nodes().Patch(context.Background(), node.Name, types.JSONPatchType, []byte(labelPatch), metav1.PatchOptions{})
	if err != nil {
		log.Fatalln("Failed to label node!")
	}

	// Delete and Create namespace
	_, err = clientset.CoreV1().Namespaces().Get(context.Background(), constants.MERAK_COMPUTE_WORKER_NAMESPACE, metav1.GetOptions{})
	if err == nil {
		log.Println("Namespace already exists, deleting and recreating... ")
		err = clientset.CoreV1().Namespaces().Delete(context.Background(), constants.MERAK_COMPUTE_WORKER_NAMESPACE, metav1.DeleteOptions{})
		if err != nil {
			log.Fatalf("Failed to delete namespace: " + constants.MERAK_COMPUTE_WORKER_NAMESPACE)
		}
		nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: constants.MERAK_COMPUTE_WORKER_NAMESPACE}}
		clientset.CoreV1().Namespaces().Create(context.Background(), nsSpec, metav1.CreateOptions{})
	}

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
	for _, host := range in.Config.VmDeploy.Hosts {
		worker := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.WORKER_POD_PREFIX + host,
			},
			Spec: corev1.PodSpec{
				NodeSelector: map[string]string{
					constants.KUBE_NODE_LABEL_KEY: constants.KUBE_NODE_LABEL_VAL,
				},
				Containers: []corev1.Container{
					{
						Name:            constants.WORKER_POD_PREFIX + host,
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
								Name: constants.TEMPORAL_ENV, Value: temporal_addr,
							},
							{
								Name: constants.TEMPORAL_RPS_ENV, Value: rps,
							},
							{
								Name: constants.TEMPORAL_CONCURRENCY_ENV, Value: concurrency,
							},
							{
								Name: constants.TEMPORAL_TQ_ENV, Value: host,
							},
						},
					},
				},
			},
		}
		_, err := clientset.CoreV1().Pods(constants.MERAK_COMPUTE_WORKER_NAMESPACE).Create(context.Background(), worker, metav1.CreateOptions{})
		if err != nil {
			log.Fatalln("Failed to create worker pod: "+constants.WORKER_POD_PREFIX+host, err.Error())
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
