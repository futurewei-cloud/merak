package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/services"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *services.Port))
	gRPCServer := grpc.NewServer()
	pb.RegisterMerakComputeServiceServer(gRPCServer, &services.Server{})

	temporal := os.Getenv(common.TEMPORAL_ENV)
	var sb strings.Builder
	sb.WriteString(temporal)
	sb.WriteString(":")
	sb.WriteString(common.TEMPORAL_PORT)

	log.Printf("Connecting to Temporal server at %s", sb.String())
	client, err := client.NewClient(client.Options{
		HostPort: sb.String(),
	})
	if err != nil {
		log.Fatalln("ERROR: Unable to create Temporal client", err)
	}
	log.Printf("Successfully connected to Temporal!")

	log.Printf("Starting gRPC server. Listening at %v", lis.Addr())
	if err := gRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer client.Close()
}
