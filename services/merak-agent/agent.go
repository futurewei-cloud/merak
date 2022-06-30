package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-agent/handler"
	"google.golang.org/grpc"
)

func main() {

	// Start plugin
	cmd := exec.Command("/bin/bash", "/merak/tools/start-aca.sh")
	cmd.Dir = "/"
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Started ACA %d\n", cmd.Process.Pid)

	// Start gRPC Server
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *handler.Port))
	if err != nil {
		log.Fatalln("ERROR: Failed to listen", err)
	}
	gRPCServer := grpc.NewServer()
	pb.RegisterMerakAgentServiceServer(gRPCServer, &handler.Server{})
	log.Printf("Starting gRPC server. Listening at %v", lis.Addr())
	if err := gRPCServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
