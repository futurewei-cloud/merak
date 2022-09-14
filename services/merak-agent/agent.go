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
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-agent/handler"
	"google.golang.org/grpc"
)

var (
	Port = flag.Int("port", constants.AGENT_GRPC_SERVER_PORT, "The server port")
)

func main() {

	if len(os.Args) < 3 {
		log.Fatal("Not enough arguments")
	}
	remote_server := os.Args[1]
	if net.ParseIP(remote_server) == nil {
		log.Fatalf("Invalid IP address %s\n", remote_server)
	}
	handler.RemoteServer = remote_server
	remote_port := os.Args[2]
	remote_port_int, err := strconv.Atoi(os.Args[2])

	if err != nil {
		log.Fatalf("Port: %d is not a valid number!\n", remote_port_int)
	}
	if remote_port_int > constants.MAX_PORT || remote_port_int < constants.MIN_PORT {
		log.Fatalf("Port: %d is not within a valid range!\n", remote_port_int)
	}
	cmdString := "service rsyslog restart && /etc/init.d/openvswitch-switch restart && /merak-bin/AlcorControlAgent -d -a " + remote_server + " -p " + remote_port
	log.Println("Executing command " + cmdString)
	// Start plugin
	cmd := exec.Command("bash", "-c", cmdString)
	cmd.Dir = "/"
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Started ACA %d\n", cmd.Process.Pid)

	// Start gRPC Server
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *Port))
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
