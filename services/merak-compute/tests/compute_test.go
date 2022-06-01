package test

import (
	"context"
	"log"
	"net"
	"testing"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-compute/grpc/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterMerakComputeServiceServer(s, &service.Server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestGrpc(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewMerakComputeServiceClient(conn)
	resp, err := client.ComputeHandler(ctx, &pb.InternalComputeConfigInfo{})
	if err != nil {
		t.Fatalf("Compute Handler failed: %v", err)
	}
	resp, err = client.TestHandler(ctx, &pb.InternalComputeConfigInfo{})
	if err != nil {
		t.Fatalf("Test Handler failed: %v", err)
	}
	log.Printf("Response: %+v", resp)
}
