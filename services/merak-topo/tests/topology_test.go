package test

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

func TestGrpc(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewMerakTopologyServiceClient(conn)

	// message InternalVNicInfo {
	// 	OperationType operation_type = 1;
	// 	string id = 2;
	// 	string name = 3;
	// 	string ip = 4;
	// }

	// message InternalVNodeInfo {
	// 	OperationType operation_type = 1;
	// 	string id = 2;
	// 	string name = 3;
	// 	VNodeType type = 4;
	// 	repeated InternalVNicInfo vnics = 5;
	// }

	// message InternalVLinkInfo {
	// 	OperationType operation_type = 1;
	// 	string id = 2;
	// 	string name = 3;
	// 	string src = 4;
	// 	string dst = 5;
	// }

	node0 := pb.InternalNodeInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "0",
		Name:          "node_0",
		Type:          pb.TopologyType_SINGLE,
		Nics:          "node0-eth1",
		Nics:          "node0-eth2",
	}
	node1 := pb.InternalNodeInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "1",
		Name:          "node_1",
		Type:          pb.TopologyType_SINGLE,
		Nics:          "node1-eth1",
		Nics:          "node1-eth2",
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

	topologyConfig := pb.InternalTopologyConfiguration{
		FormatVersion:  1,
		RevisionNumber: 1,
		RequestId:      "test",
		topologyId:     "test",
		MessageType:    pb.MessageType_FULL,
		Nodes:          []*pb.InternalVNodeInfo{node0, node1},
		Links:          []*pb.InternalVLinkInfo{link0, link1},
		ExtraInfo:      &pb.InternalTopologyExtraInfo{Info: "test"},
	}

	topology_info := pb.InternalTopologyInfo{
		OperationType:                 pb.OperationType_CREATE,
		InternalTopologyConfiguration: topologyConfig,
	}

	resp, err := client.TopologyHandler(ctx, &topology_info)
	if err != nil {
		t.Fatalf("Topology Handler failed: %v", err)
	}
	resp, err = client.TestHandler(ctx, &pb.InternalToplogyInfo{})
	if err != nil {
		t.Fatalf("Test Handler failed: %v", err)
	}
	log.Printf("Response: %+v", resp)
}
