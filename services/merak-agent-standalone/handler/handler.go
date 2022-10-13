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
package handler

import (
	"context"
	"errors"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
)

type Server struct {
	pb.UnimplementedMerakAgentServiceServer
}

var RemoteServer string

func (s *Server) PortHandler(ctx context.Context, in *pb.InternalPortConfig) (*pb.AgentReturnInfo, error) {
	log.Println("Received on PortHandler", in)

	// Parse input d
	switch op := in.OperationType; op {
	case common_pb.OperationType_CREATE:
		log.Println("Operation Create")
		return caseCreate(ctx, in)

	case common_pb.OperationType_UPDATE:

		log.Println("Update Unimplemented")
		return &pb.AgentReturnInfo{
			ReturnMessage: "Update Unimplemented",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port:          nil,
		}, errors.New("update unimplemented")

	case common_pb.OperationType_DELETE:

		log.Println("Operation Delete")
		return caseDelete(ctx, in)

	default:
		log.Println("Unknown Operation")
		return &pb.AgentReturnInfo{
			ReturnMessage: "Unknown Operation",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, errors.New("unknown operation")
	}
}
