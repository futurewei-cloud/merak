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
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	merakEvm "github.com/futurewei-cloud/merak/services/merak-agent/evm"
)

func caseDelete(ctx context.Context, in *pb.InternalPortConfig) (*pb.AgentReturnInfo, error) {

	evm, err := merakEvm.NewEvm(in.Name, in.Ip, in.Mac, in.Remoteid, in.Deviceid, in.Cidr, in.Gw, common_pb.Status_DELETING)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Invalid info for Delete EVM",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	err = evm.DeletePort(in, RemoteServer)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Delete Port request to Alcor Failed!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	err = evm.DeleteNamespace()
	if err != nil {
		log.Println("Namespace deletion failed!")
		return &pb.AgentReturnInfo{
			ReturnMessage: "Namespace deletion failed!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	err = evm.DeleteBridge()
	if err != nil {
		log.Println("Bridge deletion failed!")
		return &pb.AgentReturnInfo{
			ReturnMessage: "Bridge deletion failed!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	err = evm.DeleteDevice()
	if err != nil {
		log.Println("Failed to delete tap")
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to delete tap",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	return &pb.AgentReturnInfo{
		ReturnMessage: "Delete Success!",
		ReturnCode:    common_pb.ReturnCode_OK,
	}, nil

}
