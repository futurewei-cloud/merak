package main

import (
	"flag"
	"fmt"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-network/grpc/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	flag.Parse()
	//lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *service.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	// pb.RegisterGreeterServer(s, &server{})
	pb.RegisterMerakNetworkServiceServer(s, &service.Server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
