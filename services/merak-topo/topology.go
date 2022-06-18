package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	"github.com/futurewei-cloud/merak/services/merak-topo/grpc/service"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()

	//grpc server init check
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *service.Port))
	if err != nil {
		log.Fatalln("ERROR: Failed to listen", err)
	}
	gRPCServer := grpc.NewServer()
	pb.RegisterMerakTopologyServiceServer(gRPCServer, &service.Server{})

	log.Printf("Starting gRPC server. Listening at %v", lis.Addr())
	if err := gRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	//grpc client

	// redis init check-- pingpong test
	err = database.ConnectDatabase()

}
