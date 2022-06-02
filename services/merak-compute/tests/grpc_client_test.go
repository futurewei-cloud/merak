package test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGrpcClient(t *testing.T) {
	var compute_address strings.Builder
	compute_address.WriteString(constants.COMPUTE_GRPC_SERVER_ADDRESS)
	compute_address.WriteString(":")
	compute_address.WriteString(strconv.Itoa(constants.COMPUTE_GRPC_SERVER_PORT))

	conn, err := grpc.Dial(compute_address.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial gRPC server address!: %v", err)
	}
	client := pb.NewMerakComputeServiceClient(conn)
	resp, err := client.ComputeHandler(context.Background(), &pb.InternalComputeConfigInfo{})
	if err != nil {
		t.Fatalf("Compute Handler failed: %v", err)
	}
	resp, err = client.TestHandler(context.Background(), &pb.InternalComputeConfigInfo{})
	if err != nil {
		t.Fatalf("Test Handler failed: %v", err)
	}
	t.Logf("Response: %+v", resp)
	defer conn.Close()
}
