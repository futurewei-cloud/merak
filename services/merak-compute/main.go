package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
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

	client, err := client.NewClient(client.Options{
		HostPort: "localhost:8081",
	})
	if err != nil {
		log.Fatalln("ERROR: Unable to create Temporal client", err)
	}

	defer client.Close()

}
