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
	"errors"
	"log"
	"net"
	"testing"
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	compute_pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	network_pb "github.com/futurewei-cloud/merak/api/proto/v1/network"
	topology_pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
	"github.com/futurewei-cloud/merak/services/scenario-manager/utils"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type mockMerakTopologyServiceServer struct {
	topology_pb.UnimplementedMerakTopologyServiceServer
}

type mockMerakNetworkServiceServer struct {
	network_pb.UnimplementedMerakNetworkServiceServer
}

type mockMerakComputeServiceServer struct {
	compute_pb.UnimplementedMerakComputeServiceServer
}

func (*mockMerakTopologyServiceServer) TopologyHandler(ctx context.Context, req *topology_pb.InternalTopologyInfo) (*topology_pb.ReturnTopologyMessage, error) {
	if req.Config.GetRequestId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request id cannot be empty %v", req.Config.GetRequestId())
	}

	compute := &pb.InternalComputeInfo{Id: "1", Name: "compute1", DatapathIp: "10.0.0.1", Mac: "ff:ff:ff:ff:ff", Veth: "eth1"}
	var computes []*pb.InternalComputeInfo
	computes = append(computes, compute)

	return &topology_pb.ReturnTopologyMessage{ReturnCode: pb.ReturnCode_OK, ReturnMessage: "Topology protobuf message received", ComputeNodes: computes}, nil
}

func topologyDialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()

	topology_pb.RegisterMerakTopologyServiceServer(server, &mockMerakTopologyServiceServer{})
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
	topoConfig := &topology_pb.InternalTopologyConfiguration{
		RequestId: utils.GenUUID(),
	}
	topoInfo := &topology_pb.InternalTopologyInfo{
		Config: topoConfig,
	}
	//compute := &pb.InternalComputeInfo{Id: "compute1", Name: "compute1", Ip: "10.244.0.1", Mac: "xx.xx.xx.xx.xx", Veth: "eth0", Status: pb.Status_READY}
	//var computes []*pb.InternalComputeInfo
	//computes = append(computes, compute)
	topoRet := &topology_pb.ReturnTopologyMessage{
		ReturnCode:    pb.ReturnCode_OK,
		ReturnMessage: "Topology protobuf message received",
		//ComputeNodes:  computes,
	}

	tests := []struct {
		name          string
		data          *topology_pb.InternalTopologyInfo
		response      *topology_pb.ReturnTopologyMessage
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
			log.Printf("responseTopo: %s", proto.MarshalTextString(res))
		})
	}
}

func (*mockMerakNetworkServiceServer) NetConfigHandler(ctx context.Context, req *network_pb.InternalNetConfigInfo) (*network_pb.ReturnNetworkMessage, error) {
	if req.Config.GetRequestId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request id cannot be empty %v", req.Config.GetRequestId())
	}

	subnet := &pb.InternalSubnetInfo{SubnetId: "1", SubnetCidr: "10.0.0.0/24", SubnetGw: "10.0.0.1"}
	var subnets []*pb.InternalSubnetInfo
	subnets = append(subnets, subnet)
	vpc := &pb.InternalVpcInfo{VpcId: "1", TenantId: "1", ProjectId: "1", Subnets: subnets}
	var vpcs []*pb.InternalVpcInfo
	vpcs = append(vpcs, vpc)

	return &network_pb.ReturnNetworkMessage{ReturnCode: pb.ReturnCode_OK, ReturnMessage: "Topology protobuf message received", Vpcs: vpcs}, nil
}

func networkDialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()

	network_pb.RegisterMerakNetworkServiceServer(server, &mockMerakNetworkServiceServer{})
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
	netConfig := &network_pb.InternalNetConfigConfiguration{
		RequestId: utils.GenUUID(),
	}
	netconfInfo := &network_pb.InternalNetConfigInfo{
		Config: netConfig,
	}
	netconfRet := &network_pb.ReturnNetworkMessage{
		ReturnCode:    pb.ReturnCode_OK,
		ReturnMessage: "Network protobuf message received",
	}

	tests := []struct {
		name          string
		data          *network_pb.InternalNetConfigInfo
		response      *network_pb.ReturnNetworkMessage
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

func (*mockMerakComputeServiceServer) ComputeHandler(ctx context.Context, req *compute_pb.InternalComputeConfigInfo) (*compute_pb.ReturnComputeMessage, error) {
	if req.Config.GetRequestId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request id cannot be empty %v", req.Config.GetRequestId())
	}

	return &compute_pb.ReturnComputeMessage{ReturnCode: pb.ReturnCode_OK, ReturnMessage: "Compute protobuf message received"}, nil
}

func computeDialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()

	compute_pb.RegisterMerakComputeServiceServer(server, &mockMerakComputeServiceServer{})
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
	computeConf := &compute_pb.InternalComputeConfiguration{
		RequestId: utils.GenUUID(),
	}
	computeInfo := &compute_pb.InternalComputeConfigInfo{
		Config: computeConf,
	}
	computeRet := &pb.ReturnMessage{
		ReturnCode:    pb.ReturnCode_OK,
		ReturnMessage: "Compute protobuf message received",
	}

	tests := []struct {
		name          string
		data          *compute_pb.InternalComputeConfigInfo
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
