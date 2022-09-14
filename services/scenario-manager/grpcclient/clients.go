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

package grpcclient

import (
	"context"
	"fmt"
	"strconv"
	"time"

	compute_pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	network_pb "github.com/futurewei-cloud/merak/api/proto/v1/network"
	topology_pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/scenario-manager/logger"
	"github.com/futurewei-cloud/merak/services/scenario-manager/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	conn    *grpc.ClientConn
	timeout time.Duration
}

func NewGrpcClient(conn *grpc.ClientConn, timeout time.Duration) GrpcClient {
	return GrpcClient{
		conn:    conn,
		timeout: timeout,
	}
}

func TopologyClient(topopb *topology_pb.InternalTopologyInfo) (*topology_pb.ReturnTopologyMessage, error) {
	var conn *grpc.ClientConn

	addr := constants.TOPLOGY_GRPC_SERVER_ADDRESS + ":" + strconv.Itoa(constants.TOPLOGY_GRPC_SERVER_PORT)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(500*1024*1024), grpc.MaxCallSendMsgSize(500*1024*1024)))

	if err != nil {
		logger.Log.Errorf("can not connect to %s", err)
		return nil, fmt.Errorf("cannot connect to merak-topology grpc server: %s", err)
	}
	defer conn.Close()

	response, err := NewGrpcClient(conn, time.Second*time.Duration(utils.GetGrpcTimeout())).TopologyHandler(context.Background(), topopb)

	if err != nil {
		logger.Log.Errorf("error return from grpc server: %s", err)
		return nil, fmt.Errorf("error return from grpc server: %s", err.Error())
	}

	return response, nil
}

func (g GrpcClient) TopologyHandler(ctx context.Context, topopb *topology_pb.InternalTopologyInfo) (*topology_pb.ReturnTopologyMessage, error) {
	client := topology_pb.NewMerakTopologyServiceClient(g.conn)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(g.timeout))
	defer cancel()

	response, err := client.TopologyHandler(ctx, topopb)

	if err != nil {
		logger.Log.Errorf("Error when calling Merak-Topology: %s", err.Error())
		return nil, fmt.Errorf("error when calling merak-topology grpc server: %s", err.Error())
	}
	logger.Log.Debugf("Response from Merak-topology grpc server: %s", response.GetReturnMessage())

	return response, nil
}

func NetworkClient(netconfpb *network_pb.InternalNetConfigInfo) (*network_pb.ReturnNetworkMessage, error) {
	var conn *grpc.ClientConn

	addr := constants.NETWORK_GRPC_SERVER_ADDRESS + ":" + strconv.Itoa(constants.NETWORK_GRPC_SERVER_PORT)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(500*1024*1024), grpc.MaxCallSendMsgSize(500*1024*1024)))

	if err != nil {
		logger.Log.Errorf("can not connect to %s", err)
		return nil, fmt.Errorf("cannot connect to merak-network grpc server: %s", err)
	}
	defer conn.Close()

	response, err := NewGrpcClient(conn, time.Second*time.Duration(utils.GetGrpcTimeout())).NetConfigHandler(context.Background(), netconfpb)

	if err != nil {
		logger.Log.Errorf("error return from grpc server: %s", err)
		return nil, fmt.Errorf("error return from grpc server: %s", err.Error())
	}

	return response, nil
}

func (g GrpcClient) NetConfigHandler(ctx context.Context, netconfpb *network_pb.InternalNetConfigInfo) (*network_pb.ReturnNetworkMessage, error) {
	client := network_pb.NewMerakNetworkServiceClient(g.conn)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(g.timeout))
	defer cancel()

	response, err := client.NetConfigHandler(ctx, netconfpb)

	if err != nil {
		logger.Log.Errorf("Error when calling Merak-Network: %s", err)
		return nil, fmt.Errorf("error when calling merak-network grpc server: %s", err)
	}
	logger.Log.Debugf("Response from Merak-Network grpc server: %s", response.GetReturnMessage())

	return response, nil
}

func ComputeClient(computepb *compute_pb.InternalComputeConfigInfo) (*compute_pb.ReturnComputeMessage, error) {
	var conn *grpc.ClientConn

	addr := constants.COMPUTE_GRPC_SERVER_ADDRESS + ":" + strconv.Itoa(constants.COMPUTE_GRPC_SERVER_PORT)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(500*1024*1024), grpc.MaxCallSendMsgSize(500*1024*1024)))
	if err != nil {
		logger.Log.Errorf("can not connect to %s", err)
		return nil, fmt.Errorf("cannot connect to merak-compute grpc server: %s", err)
	}
	defer conn.Close()

	response, err := NewGrpcClient(conn, time.Second*time.Duration(utils.GetGrpcTimeout())).ComputeHandler(context.Background(), computepb)

	if err != nil {
		logger.Log.Errorf("error return from grpc server: %s", err)
		return nil, fmt.Errorf("error connecting to grpc server: %s", err.Error())
	}

	return response, nil
}

func (g GrpcClient) ComputeHandler(ctx context.Context, computepb *compute_pb.InternalComputeConfigInfo) (*compute_pb.ReturnComputeMessage, error) {
	client := compute_pb.NewMerakComputeServiceClient(g.conn)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(g.timeout))
	defer cancel()

	response, err := client.ComputeHandler(ctx, computepb)

	if err != nil {
		logger.Log.Errorf("Error when calling Merak-Compute: %s", err)
		return nil, fmt.Errorf("error when calling merak-compute grpc server: %s", err)
	}
	logger.Log.Debugf("Response from Merak-Compute grpc server: %s", response.GetReturnMessage())

	return response, nil
}
