package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:40053", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	testInternalNetConfigConfiguration := pb.InternalNetConfigConfiguration{
		FormatVersion:  0,
		RevisionNumber: 0,
		RequestId:      "InternalNetConfigConfigurationRequestId",
		NetconfigId:    "InternalNetConfigConfigurationNetconfigId",
	}
	testInternalNetConfigInfo := pb.InternalNetConfigInfo{
		OperationType: pb.OperationType_DELETE,
		Config:        &testInternalNetConfigConfiguration,
	}

	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	// c := pb.NewGreeterClient(conn)
	c := pb.NewMerakNetworkServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	r, err := c.NetConfigHandler(ctx, &testInternalNetConfigInfo)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Return: %s", r)
}
