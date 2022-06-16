package grpcclient

import (
	"context"
	"fmt"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"google.golang.org/grpc"
)

func TopologyClient(topopb *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(":40052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
		return nil, fmt.Errorf("Cannot connect to merak-topology grpc server: %s", err)
	}
	defer conn.Close()

	c := pb.NewMerakTopologyServiceClient(conn)
	response, err := c.TopologyHandler(context.Background(), topopb)
	if err != nil {
		log.Fatalf("Error when calling Merak-Topology: %s", err)
		return nil, fmt.Errorf("Error when calling merak-topology grpc server: %s", err)
	}
	log.Printf("Response from server: %s", response.GetReturnMessage())

	return response, nil
}

func NetworkClient(netconfpb *pb.InternalNetConfigInfo) (*pb.ReturnNetworkMessage, error) {
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(":40053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
		return nil, fmt.Errorf("Cannot connect to merak-network grpc server: %s", err)
	}
	defer conn.Close()

	c := pb.NewMerakNetworkServiceClient(conn)
	response, err := c.InternalNetConfigConfigurationHandler(context.Background(), netconfpb)
	if err != nil {
		log.Fatalf("Error when calling Merak-Topology: %s", err)
		return nil, fmt.Errorf("Error when calling merak-topology grpc server: %s", err)
	}
	log.Printf("Response from server: %s", response.GetReturnMessage())

	return response, nil
}

func ComputeClient(computepb *pb.InternalComputeConfigInfo) (*pb.ReturnMessage, error) {
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(":40051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
		return nil, fmt.Errorf("Cannot connect to merak-compute grpc server: %s", err)
	}
	defer conn.Close()

	c := pb.NewMerakComputeServiceClient(conn)
	response, err := c.ComputeHandler(context.Background(), computepb)
	if err != nil {
		log.Fatalf("Error when calling Merak-Topology: %s", err)
		return nil, fmt.Errorf("Error when calling merak-topology grpc server: %s", err)
	}
	log.Printf("Response from server: %s", response.GetReturnMessage())

	return response, nil
}
