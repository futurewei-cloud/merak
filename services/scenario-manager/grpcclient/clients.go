package grpcclient

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
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

func TopologyClient(topopb *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	var conn *grpc.ClientConn

	addr := constants.TOPLOGY_GRPC_SERVER_ADDRESS + ":" + strconv.Itoa(constants.TOPLOGY_GRPC_SERVER_PORT)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

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

func (g GrpcClient) TopologyHandler(ctx context.Context, topopb *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	client := pb.NewMerakTopologyServiceClient(g.conn)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(g.timeout))
	defer cancel()

	response, err := client.TopologyHandler(ctx, topopb)

	if err != nil {
		logger.Log.Errorf("Error when calling Merak-Topology: %s", err.Error())
		return nil, fmt.Errorf("error when calling merak-topology grpc server: %s", err.Error())
	}
	logger.Log.Printf("Response from server: %s", response.GetReturnMessage())

	return response, nil
}

func NetworkClient(netconfpb *pb.InternalNetConfigInfo) (*pb.ReturnNetworkMessage, error) {
	var conn *grpc.ClientConn

	addr := constants.NETWORK_GRPC_SERVER_ADDRESS + ":" + strconv.Itoa(constants.NETWORK_GRPC_SERVER_PORT)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (g GrpcClient) NetConfigHandler(ctx context.Context, netconfpb *pb.InternalNetConfigInfo) (*pb.ReturnNetworkMessage, error) {
	client := pb.NewMerakNetworkServiceClient(g.conn)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(g.timeout))
	defer cancel()

	response, err := client.NetConfigHandler(ctx, netconfpb)

	if err != nil {
		logger.Log.Errorf("Error when calling Merak-Network: %s", err)
		return nil, fmt.Errorf("error when calling merak-network grpc server: %s", err)
	}
	logger.Log.Printf("Response from server: %s", response.GetReturnMessage())

	return response, nil
}

func ComputeClient(computepb *pb.InternalComputeConfigInfo) (*pb.ReturnMessage, error) {
	var conn *grpc.ClientConn

	addr := constants.COMPUTE_GRPC_SERVER_ADDRESS + ":" + strconv.Itoa(constants.COMPUTE_GRPC_SERVER_PORT)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (g GrpcClient) ComputeHandler(ctx context.Context, computepb *pb.InternalComputeConfigInfo) (*pb.ReturnMessage, error) {
	client := pb.NewMerakComputeServiceClient(g.conn)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(g.timeout))
	defer cancel()

	response, err := client.ComputeHandler(ctx, computepb)

	if err != nil {
		logger.Log.Errorf("Error when calling Merak-Compute: %s", err)
		return nil, fmt.Errorf("error when calling merak-compute grpc server: %s", err)
	}
	logger.Log.Printf("Response from server: %s", response.GetReturnMessage())

	return response, nil
}
