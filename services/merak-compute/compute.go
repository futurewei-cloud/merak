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

	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/handler"
	"github.com/go-redis/redis/v9"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

var ctx = context.Background()

func main() {
	// Connect to temporal
	temporal_address, ok := os.LookupEnv(constants.TEMPORAL_ENV)
	if !ok {
		log.Println("Temporal environment variable not set, using default address.")
		temporal_address = constants.TEMPORAL_ADDRESS
	}
	var sb strings.Builder
	sb.WriteString(temporal_address)
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(constants.TEMPORAL_PORT))
	var err error
	log.Printf("Connecting to Temporal server at %s", sb.String())

	handler.TemporalClient, err = client.NewClient(client.Options{
		HostPort: sb.String(),
	})
	if err != nil {
		log.Fatalln("ERROR: Unable to create Temporal client", err)
	}
	log.Println("Successfully connected to Temporal!")
	defer handler.TemporalClient.Close()

	//Connect to Redis
	var redisAddress strings.Builder
	redisAddress.WriteString(constants.COMPUTE_REDIS_ADDRESS)
	redisAddress.WriteString(":")
	redisAddress.WriteString(strconv.Itoa(constants.COMPUTE_REDIS_PORT))

	handler.RedisClient = *redis.NewClient(&redis.Options{
		Addr:     redisAddress.String(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err = handler.RedisClient.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		log.Fatalln("ERROR: Unable to create Redis client", err)
	}
	log.Println("Successfully connected to Redis!")
	defer handler.RedisClient.Close()

	//Start gRPC Server
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *handler.Port))
	if err != nil {
		log.Fatalln("ERROR: Failed to listen", err)
	}
	gRPCServer := grpc.NewServer()
	pb.RegisterMerakComputeServiceServer(gRPCServer, &handler.Server{})
	log.Printf("Starting gRPC server. Listening at %v", lis.Addr())
	if err := gRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
