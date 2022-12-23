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

	"net"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"

	"github.com/futurewei-cloud/merak/services/merak-topo/grpc/service"
	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
	"google.golang.org/grpc"
)

//merak-top main
func main() {
	flag.Parse()

	lis, err1 := net.Listen("tcp", fmt.Sprintf(":%d", *service.Port))
	if err1 != nil {
		utils.Logger.Fatal("Fail to listen", err1)
	}

	gRPCServer := grpc.NewServer()
	pb.RegisterMerakTopologyServiceServer(gRPCServer, &service.Server{})

	utils.Logger.Info("Starting gRPC server. Listening at %v", lis.Addr())

	err2 := gRPCServer.Serve(lis)
	if err2 != nil {
		utils.Logger.Fatal("Can not connect to gPRCServe: %v", err2)
	}

}
