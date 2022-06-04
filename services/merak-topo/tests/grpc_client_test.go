package tests

import (
	"context"
	"strconv"
	"strings"
	"testing"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"google.golang.org/grpc"
)

func TestGrpcClient(t *testing.T) {
	var topology_address strings.Builder
	// topology_address.WriteString(constants.TOPOLOGY_GRPC_SERVER_ADDRESS)
	topology_address.WriteString(":")
	topology_address.WriteString(strconv.Itoa(constants.TOPOLOGY_GRPC_SERVER_PORT))

	conn, err := grpc.Dial(topology_address.String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial gRPC server address!: %v", err)
	}
	defer conn.Close()

	client := pb.NewMerakTopologyServiceClient(conn)
	resp, err := client.TopologyHandler(context.Background(), &pb.InternalTopologyInfo{})
	if err != nil {
		t.Fatalf("Topology Handler failed: %v", err)
	}
	resp, err = client.TestHandler(context.Background(), &pb.InternalTopologyInfo{})
	if err != nil {
		t.Fatalf("Test Handler failed: %v", err)
	}
	t.Logf("Response: %+v", resp)
	defer conn.Close()
}
