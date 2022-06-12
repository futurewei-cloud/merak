package grpcclient

import (
	"context"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"google.golang.org/grpc"
)

func TopologyClient(topopb *pb.InternalTopologyInfo) error {
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(":7777", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := pb.NewMerakTopologyServiceClient(conn)
	response, err := c.TopologyHandler(context.Background(), &pb.InternalTopologyInfo{})
	if err != nil {
		log.Fatalf("Error when calling Merak-Topology: %s", err)
	}
	log.Printf("Response from server: %s", response.GetConfig())

	return nil
}
