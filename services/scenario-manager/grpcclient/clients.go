package grpcclient

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
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

	conn, err := grpc.Dial(":40052", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

	if err != nil {
		log.Fatalf("did not connect: %s", err)
		return nil, fmt.Errorf("Cannot connect to merak-topology grpc server: %s", err)
	}
	defer conn.Close()

	response, err := NewGrpcClient(conn, time.Second).TopologyHandler(context.Background(), topopb)

	if err != nil {
		return nil, fmt.Errorf("error connecting to grpc server: %s", err.Error())
	}

	return response, nil
}

func (g GrpcClient) TopologyHandler(ctx context.Context, topopb *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	client := pb.NewMerakTopologyServiceClient(g.conn)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(g.timeout))
	defer cancel()

	response, err := client.TopologyHandler(ctx, topopb)

	if err != nil {
		log.Fatalf("Error when calling Merak-Topology: %s", err.Error())
		return nil, fmt.Errorf("Error when calling merak-topology grpc server: %s", err.Error())
	}
	log.Printf("Response from server: %s", response.GetReturnMessage())

	return response, nil
}

func NetworkClient(netconfpb *pb.InternalNetConfigInfo) (*pb.ReturnNetworkMessage, error) {
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(":40053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %s", err)
		return nil, fmt.Errorf("Cannot connect to merak-network grpc server: %s", err)
	}
	defer conn.Close()

	response, err := NewGrpcClient(conn, time.Second).NetConfigHandler(context.Background(), netconfpb)

	if err != nil {
		return nil, fmt.Errorf("error connecting to grpc server: %s", err.Error())
	}

	return response, nil
}

func (g GrpcClient) NetConfigHandler(ctx context.Context, netconfpb *pb.InternalNetConfigInfo) (*pb.ReturnNetworkMessage, error) {
	client := pb.NewMerakNetworkServiceClient(g.conn)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(g.timeout))
	defer cancel()

	response, err := client.NetConfigHandler(ctx, netconfpb)

	if err != nil {
		log.Fatalf("Error when calling Merak-Network: %s", err)
		return nil, fmt.Errorf("Error when calling merak-network grpc server: %s", err)
	}
	log.Printf("Response from server: %s", response.GetReturnMessage())

	return response, nil
}

func ComputeClient(computepb *pb.InternalComputeConfigInfo) (*pb.ReturnMessage, error) {
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(":40051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %s", err)
		return nil, fmt.Errorf("Cannot connect to merak-compute grpc server: %s", err)
	}
	defer conn.Close()

	response, err := NewGrpcClient(conn, time.Second).ComputeHandler(context.Background(), computepb)

	if err != nil {
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
		log.Fatalf("Error when calling Merak-Compute: %s", err)
		return nil, fmt.Errorf("Error when calling merak-compute grpc server: %s", err)
	}
	log.Printf("Response from server: %s", response.GetReturnMessage())

	return response, nil
}
