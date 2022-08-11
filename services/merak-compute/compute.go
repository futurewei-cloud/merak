package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/grpc/service"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *service.Port))
	if err != nil {
		log.Fatalln("ERROR: Failed to listen", err)
	}
	gRPCServer := grpc.NewServer()
	pb.RegisterMerakComputeServiceServer(gRPCServer, &service.Server{})

	temporal_address, ok := os.LookupEnv(constants.TEMPORAL_ENV)
	if !ok {
		log.Println("Temporal environment variable not set, using default address.")
		temporal_address = constants.TEMPRAL_ADDRESS
	}
	var sb strings.Builder
	sb.WriteString(temporal_address)
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(constants.TEMPORAL_PORT))

	log.Printf("Connecting to Temporal server at %s", sb.String())
	service.TemporalClient, err = client.NewClient(client.Options{
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

	defer service.TemporalClient.Close()
}
