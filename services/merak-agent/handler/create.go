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

func caseCreate(ctx context.Context, in *pb.InternalPortConfig) (*pb.AgentReturnInfo, error) {
	var evm *merakEvm.AlcorEvm
	var err error
	_, ok := os.LookupEnv(constants.AGENT_MODE_ENV)
	if !ok {
		evm, err = merakEvm.CreateMinimalPort(in, RemoteServer, MerakMetrics)
		if err != nil {
			return &pb.AgentReturnInfo{
				ReturnMessage: "Create Minimal Port Failed",
				ReturnCode:    common_pb.ReturnCode_FAILED,
				Port: &pb.ReturnPortInfo{
					Status: common_pb.Status_ERROR,
				},
			}, err

		}
		err = evm.CreateDevice(MerakMetrics)
		if err != nil {
			return &pb.AgentReturnInfo{
				ReturnMessage: "ovs-vsctl command failed!",
				ReturnCode:    common_pb.ReturnCode_FAILED,
				Port: &pb.ReturnPortInfo{
					Status: common_pb.Status_ERROR,
				},
			}, err
		}
	} else {
		evm, _ = merakEvm.NewEvm(
			in.Name,
			constants.AGENT_STANDALONE_IP,
			constants.AGENT_STANDALONE_MAC,
			constants.AGENT_STANDALONE_REMOTE_ID,
			"tap"+in.Name,
			constants.AGENT_STANDALONE_CIDR,
			constants.AGENT_STANDALONE_GW,
			common_pb.Status_DEPLOYING,
		)
		err := evm.CreateStandaloneDevice(MerakMetrics)
		if err != nil {
			return &pb.AgentReturnInfo{
				ReturnMessage: "Failed to create tap",
				ReturnCode:    common_pb.ReturnCode_FAILED,
				Port: &pb.ReturnPortInfo{
					Status: common_pb.Status_ERROR,
				},
			}, err
		}
	}

	err = evm.CreateNamespace(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Namespace creation failed!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.CreateVethPair(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Inner and outer veth creation failed!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.MoveVethToNamespace(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Move veth into namespace failed!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.AssignIP(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to give inner veth IP!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.BringInnerVethUp(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed bring up inner veth!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.SetMTUProbing(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to set MTU probing!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.BringOuterVethUp(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to bring up outer veth!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.BringLoUp(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to bring up loopback!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.AssignMac(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Assign mac!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.AddGateway(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed add default gw!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.CreateBridge(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to create bridge!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.BringBridgeUp(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to bring bridge up!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.AddVethToBridge(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to add veth to bridge!",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.AddDeviceToBridge(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to add tap to bridge",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.BringBridgeUp(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to bring up bridge",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	err = evm.BringDeviceUp(MerakMetrics)
	if err != nil {
		return &pb.AgentReturnInfo{
			ReturnMessage: "Failed to bring up tap device",
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Port: &pb.ReturnPortInfo{
				Status: common_pb.Status_ERROR,
			},
		}, err
	}

	log.Println("Successfully created devices for evm ", evm.GetName())

	_, ok = os.LookupEnv(constants.AGENT_MODE_ENV)
	if !ok {
		err = evm.UpdatePort(in, RemoteServer, MerakMetrics)
		if err != nil {
			return &pb.AgentReturnInfo{
				ReturnMessage: "Failed to update port",
				ReturnCode:    common_pb.ReturnCode_FAILED,
				Port: &pb.ReturnPortInfo{
					Status: common_pb.Status_ERROR,
				},
			}, err
		}
	}

	return &pb.AgentReturnInfo{
		ReturnMessage: "Create Success",
		ReturnCode:    common_pb.ReturnCode_OK,
		Port: &pb.ReturnPortInfo{
			Ip:       evm.GetIP(),
			Deviceid: evm.GetDeviceId(),
			Remoteid: evm.GetRemoteId(),
			Status:   common_pb.Status_DONE,
		},
	}, nil
}
