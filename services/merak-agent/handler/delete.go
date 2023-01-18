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
	"os"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	constants "github.com/futurewei-cloud/merak/services/common"
	merakEvm "github.com/futurewei-cloud/merak/services/merak-agent/evm"
)

func caseDelete(ctx context.Context, in *pb.InternalPortConfig, deletePortUrl string) (*pb.AgentReturnInfo, error) {
	evm, err := merakEvm.NewEvm(
		in.Name,
		constants.AGENT_STANDALONE_IP,
		constants.AGENT_STANDALONE_MAC,
		in.Remoteid,
		in.Deviceid,
		constants.AGENT_STANDALONE_CIDR,
		constants.AGENT_STANDALONE_GW,
		common_pb.Status_DELETING)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Invalid info for Delete EVM",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}
	err = evm.MoveDeviceToRootNetns(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to move device to root namespace",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	val, ok := os.LookupEnv(constants.MODE_ENV)
	if !ok {
		val = constants.MODE_ALCOR
	}
	MerakLogger.Info("Executing in mode " + val)
	if val == constants.MODE_ALCOR {
		err = merakEvm.DeletePort(deletePortUrl+evm.GetRemoteId(), in, MerakMetrics, evm)
		if err != nil {
			return &pb.AgentReturnInfo{
				ReturnMessage: "Delete Port request to Alcor Failed!",
				ReturnCode:    common_pb.ReturnCode_FAILED,
			}, err
		}
	}

	err = evm.DeleteNamespace(MerakMetrics)
	if err != nil {
		log.Println("Namespace deletion failed!")
		return &pb.AgentReturnInfo{
			ReturnMessage: "Namespace deletion failed!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}

	if val == constants.MODE_ALCOR {
		err = evm.DeleteDevice(MerakMetrics)
		if err != nil {
			log.Println("Failed to delete tap")
			return &pb.AgentReturnInfo{
				ReturnMessage: "Failed to delete tap",
				ReturnCode:    common_pb.ReturnCode_FAILED,
			}, err
		}
	} else {
		err = evm.DeleteStandaloneDevice(MerakMetrics)
		if err != nil {
			log.Println("Failed to delete tap")
			return &pb.AgentReturnInfo{
				ReturnMessage: "Failed to delete tap",
				ReturnCode:    common_pb.ReturnCode_FAILED,
			}, err
		}
	}

	log.Println("Successfully deleted devices for evm ", evm.GetName())
	return &pb.AgentReturnInfo{
		ReturnMessage: "Delete Success!",
		ReturnCode:    common_pb.ReturnCode_OK,
	}, nil
}
