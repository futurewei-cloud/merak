package grpcclient

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/scenario-manager/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type mockMerakTopologyServiceServer struct {
	pb.UnimplementedMerakTopologyServiceServer
}

type mockMerakNetworkServiceServer struct {
	pb.UnimplementedMerakNetworkServiceServer
}

type mockMerakComputeServiceServer struct {
	pb.UnimplementedMerakComputeServiceServer
}

func (*mockMerakTopologyServiceServer) TopologyHandler(ctx context.Context, req *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	if req.Config.GetRequestId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request id cannot be empty %v", req.Config.GetRequestId())
	}

	compute := &pb.InternalComputeInfo{Id: "1", Name: "compute1", Ip: "10.0.0.1", Mac: "ff:ff:ff:ff:ff", Veth: "eth1"}
	var computes []*pb.InternalComputeInfo
	computes = append(computes, compute)

	return &pb.ReturnTopologyMessage{ReturnCode: pb.ReturnCode_OK, ReturnMessage: "Topology protobuf message received", ComputeNodes: computes}, nil
}

func topologyDialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()

	pb.RegisterMerakTopologyServiceServer(server, &mockMerakTopologyServiceServer{})
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestTopologyClient(t *testing.T) {
	topoConfig := &pb.InternalTopologyConfiguration{
		RequestId: utils.GenUUID(),
	}
	topoInfo := &pb.InternalTopologyInfo{
		Config: topoConfig,
	}
	topoRet := &pb.ReturnTopologyMessage{
		ReturnCode:    pb.ReturnCode_OK,
		ReturnMessage: "Topology protobuf message received",
	}

	tests := []struct {
		name          string
		data          *pb.InternalTopologyInfo
		response      *pb.ReturnTopologyMessage
		expectedError bool
		err           error
	}{
		{
			"Test topology client",
			topoInfo,
			topoRet,
			false,
			nil,
		},
	}

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(topologyDialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := NewGrpcClient(conn, time.Second).TopologyHandler(context.Background(), test.data)

			if res.ReturnCode != test.response.ReturnCode {
				t.Error("error: expected", test.expectedError, "received", true)
			}
			if err != nil && errors.Is(err, test.err) {
				t.Error("error: expected", test.err, "received", err)
			}
		})
	}
}

func (*mockMerakNetworkServiceServer) NetConfigHandler(ctx context.Context, req *pb.InternalNetConfigInfo) (*pb.ReturnNetworkMessage, error) {
	if req.Config.GetRequestId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request id cannot be empty %v", req.Config.GetRequestId())
	}

	subnet := &pb.InternalSubnetInfo{SubnetId: "1", SubnetCidr: "10.0.0.0/24", SubnetGw: "10.0.0.1"}
	var subnets []*pb.InternalSubnetInfo
	subnets = append(subnets, subnet)
	vpc := &pb.InternalVpcInfo{VpcId: "1", TenantId: "1", ProjectId: "1", Subnets: subnets}
	var vpcs []*pb.InternalVpcInfo
	vpcs = append(vpcs, vpc)

	return &pb.ReturnNetworkMessage{ReturnCode: pb.ReturnCode_OK, ReturnMessage: "Topology protobuf message received", Vpcs: vpcs}, nil
}

func networkDialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()

	pb.RegisterMerakNetworkServiceServer(server, &mockMerakNetworkServiceServer{})
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestNetworkClient(t *testing.T) {
	netConfig := &pb.InternalNetConfigConfiguration{
		RequestId: utils.GenUUID(),
	}
	netconfInfo := &pb.InternalNetConfigInfo{
		Config: netConfig,
	}
	netconfRet := &pb.ReturnNetworkMessage{
		ReturnCode:    pb.ReturnCode_OK,
		ReturnMessage: "Network protobuf message received",
	}

	tests := []struct {
		name          string
		data          *pb.InternalNetConfigInfo
		response      *pb.ReturnNetworkMessage
		expectedError bool
		err           error
	}{
		{
			"Test network grpc client",
			netconfInfo,
			netconfRet,
			false,
			nil,
		},
	}

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(networkDialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := NewGrpcClient(conn, time.Second).NetConfigHandler(context.Background(), test.data)

			if res.ReturnCode != test.response.ReturnCode {
				t.Error("error: expected", test.expectedError, "received", true)
			}
			if err != nil && errors.Is(err, test.err) {
				t.Error("error: expected", test.err, "received", err)
			}
		})
	}
}

func (*mockMerakComputeServiceServer) ComputeHandler(ctx context.Context, req *pb.InternalComputeConfigInfo) (*pb.ReturnMessage, error) {
	if req.Config.GetRequestId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request id cannot be empty %v", req.Config.GetRequestId())
	}

	return &pb.ReturnMessage{ReturnCode: pb.ReturnCode_OK, ReturnMessage: "Compute protobuf message received"}, nil
}

func computeDialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()

	pb.RegisterMerakComputeServiceServer(server, &mockMerakComputeServiceServer{})
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestComputeClient(t *testing.T) {
	computeConf := &pb.InternalComputeConfiguration{
		RequestId: utils.GenUUID(),
	}
	computeInfo := &pb.InternalComputeConfigInfo{
		Config: computeConf,
	}
	computeRet := &pb.ReturnMessage{
		ReturnCode:    pb.ReturnCode_OK,
		ReturnMessage: "Compute protobuf message received",
	}

	tests := []struct {
		name          string
		data          *pb.InternalComputeConfigInfo
		response      *pb.ReturnMessage
		expectedError bool
		err           error
	}{
		{
			"Test compute grpc client",
			computeInfo,
			computeRet,
			false,
			nil,
		},
	}

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(computeDialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := NewGrpcClient(conn, time.Second).ComputeHandler(context.Background(), test.data)

			if res.ReturnCode != test.response.ReturnCode {
				t.Error("error: expected", test.expectedError, "received", true)
			}
			if err != nil && errors.Is(err, test.err) {
				t.Error("error: expected", test.err, "received", err)
			}
		})
	}
}
