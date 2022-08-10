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
package tests

import (
	"context"
	"log"
	"net"
	"testing"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-topo/grpc/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// Test cases for operation INFO, CREATE, DELETE, UPDATE

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterMerakTopologyServiceServer(s, &service.Server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestGRPC(t *testing.T) {
	// Test case setup
	node0 := pb.InternalVNodeInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "0",
		Name:          "proj1-topo1-vnode0",
		Type:          pb.VNodeType_VHOST,
		Vnics:         []*pb.InternalVNicInfo{},
	}
	node1 := pb.InternalVNodeInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "1",
		Name:          "proj1-topo1-vnode1",
		Type:          pb.VNodeType_VSWITCH,
		Vnics:         []*pb.InternalVNicInfo{},
	}

	link0 := pb.InternalVLinkInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "0",
		Name:          "link_0",
		Src:           "10.0.0.1",
		Dst:           "10.0.0.2",
	}

	link1 := pb.InternalVLinkInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "1",
		Name:          "link_1",
		Src:           "10.0.0.2",
		Dst:           "10.0.0.3",
	}

	topologyConfig_c1 := pb.InternalTopologyConfiguration{
		FormatVersion:  1,
		RevisionNumber: 1,
		RequestId:      "proj1-topo1-info-test",
		TopologyId:     "proj1-topo1",
		MessageType:    pb.MessageType_FULL,
		Vnodes:         []*pb.InternalVNodeInfo{&node0, &node1},
		Vlinks:         []*pb.InternalVLinkInfo{&link0, &link1},
		ExtraInfo:      &pb.InternalTopologyExtraInfo{Info: "info test"},
	}

	topologyConfig_c2 := pb.InternalTopologyConfiguration{
		FormatVersion:  1,
		RevisionNumber: 1,
		RequestId:      "proj1-topo2-create-test",
		TopologyId:     "proj1-topo2",
		MessageType:    pb.MessageType_FULL,
		Vnodes:         []*pb.InternalVNodeInfo{},
		Vlinks:         []*pb.InternalVLinkInfo{},
		ExtraInfo:      &pb.InternalTopologyExtraInfo{Info: "create test"},
	}
	// Test cases for INFO, CREATE, DELETE, UPDATE
	topology_info := pb.InternalTopologyInfo{
		OperationType: pb.OperationType_INFO,
		Config:        &topologyConfig_c1,
	}

	topology_create := pb.InternalTopologyInfo{
		OperationType: pb.OperationType_CREATE,
		Config:        &topologyConfig_c2,
	}

	// gRPC Setup
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewMerakTopologyServiceClient(conn)

	// Run Test for INFO
	resp1, err1 := client.TopologyHandler(ctx, &topology_info)
	if err1 != nil {
		t.Fatalf("Topology Handler failed: %v", err1)
	}
	log.Printf("TopologyHandler Response: %+v", resp1)

	// Run Test for CREATE
	resp2, err2 := client.TopologyHandler(ctx, &topology_create)
	if err2 != nil {
		t.Fatalf("Topology Handler failed: %v", err2)
	}
	log.Printf("TopologyHandler Response: %+v", resp2)

	defer conn.Close()

}
