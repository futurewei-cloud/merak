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
	"github.com/futurewei-cloud/merak/services/merak-compute/handler"
	"github.com/go-redis/redis/v9"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
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
		PoolSize:    1000,
		PoolTimeout: 60,
	})

	err = handler.RedisClient.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		log.Fatalln("ERROR: Unable to create Redis client", err)
	}
	log.Println("Successfully connected to Redis!")
	defer handler.RedisClient.Close()

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
