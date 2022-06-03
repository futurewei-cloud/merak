package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-topo/grpc/service"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *service.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)

	}

	grpcServer := grpc.NewServer()
	pb.RegisterMerakTopologyServiceServer(gRPCServer, &service.Server{})

	temporal_address := os.Getenv(constants.TEMPORAL_ENV)

	var sb strings.Builder
	sb.WriteString(temporal_address)
	sb.WriteString(":")
	sb.WriteString(constants.TEMPORAL_PORT)

	log.Printf("Connecting to Temporal server at %s", sb.String())
	service.TemporalClient, err = client.NewClient(client.Options{
		HostPort: sb.String(),
	})

	if err != nil {
		log.Fatalln("ERROR: unable to create Temporal client", err)
	}

	log.Printf("Successfully connected to Temporal!")

	log.Printf("Starting gRPC server and listening at %v", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC server: %v.", err)
	}

	defer service.TemporalClient.Close()

}
