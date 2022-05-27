package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/src/common"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", common.COMPUTE_GRPC_SERVER_PORT, "The server port")
)

type server struct {
	pb.UnimplementedMerakComputeServiceServer
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	gRPCServer := grpc.NewServer()
	pb.RegisterMerakComputeServiceServer(gRPCServer, &server{})
	log.Printf("gRPC server listening at %v", lis.Addr())

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

	log.Printf("Starting gRPC Server!")
	if err := gRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer client.Close()
}
