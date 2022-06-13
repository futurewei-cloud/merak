package grpcclient

import (
	"context"
	"fmt"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"google.golang.org/grpc"
)

func TopologyClient(topopb *pb.InternalTopologyInfo) (*pb.ReturnMessage, error) {
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(":7777", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
		return nil, fmt.Errorf("Cannot connect to merak-topology grpc server: %s", err)
	}
	defer conn.Close()

	c := pb.NewMerakTopologyServiceClient(conn)
	response, err := c.TopologyHandler(context.Background(), &pb.InternalTopologyInfo{})
	if err != nil {
		log.Fatalf("Error when calling Merak-Topology: %s", err)
		return nil, fmt.Errorf("Error when calling merak-topology grpc server: %s", err)
	}
	//log.Printf("Response from server: %s", response.GetReturnMessage())

	return response, nil
}
