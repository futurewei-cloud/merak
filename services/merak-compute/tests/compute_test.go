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

	// message InternalPodInfo {
	// 	OperationType operation_type = 1;
	// 	string id = 2;
	// 	string name = 3;
	// 	string ip = 4;
	// }

	// message InternalVMInfo {
	// 	OperationType operation_type = 1;
	// 	VMDeployType deploy_type = 2;
	// 	repeated string vpcs = 3;
	// 	repeated string subnets = 4;
	// 	repeated string secgroups = 5;
	// 	uint32 num_port_per_vm = 6;
	// 	repeated InternalVMPod deploy_method = 7;
	// 	string service_url = 8;
	// }

	// message InternalVMPod {
	// 	OperationType operation_type = 1;
	// 	string pod_ip = 2;
	// 	uint32 num_of_vm = 3;
	// 	repeated string subnets = 4;
	// }

	pod0 := pb.InternalPodInfo {
		OperationType: pb.OperationType_CREATE,
		Id: "0",
		Name: "pod_name",
		Ip: "10.0.0.2",
	}
	pod1 := pb.InternalPodInfo {
		OperationType: pb.OperationType_CREATE,
		Id: "1",
		Name: "pod_name",
		Ip: "10.0.0.3",
	}
	pod2 := pb.InternalPodInfo {
		OperationType: pb.OperationType_CREATE,
		Id: "2",
		Name: "pod_name",
		Ip: "10.0.0.4",
	}
	// vms := pb.InternalVMInfo {
	// 	OperationType: pb.OperationType_CREATE,
	// 	DeployType: pb.VMDeployType_RANDOM,
	// 	Vpcs: []string{"vpc0", "vpc1"},
	// 	Subnets: []string{"subnet0", "subnet1"},
	// 	Secgroups: []string{"sg0", "sg1"},
	// 	NumPortPerVm: 2,
	// 	DeployMethod: []depl,

	// }
	computeConfig := pb.InternalComputeConfiguration{
		FormatVersion: 1,
		RevisionNumber: 1,
		RequestId: "test",
		ComputeConfigId: "test",
		MessageType: pb.MessageType_FULL,
		Pods: []*pb.InternalPodInfo{pod0, pod1, pod2},
		Vms: []*pb.InternalVMInfo{},
		VMScheduleType: pb.VMScheduleType_SEQUENTIAL,
		ExtraInfo: &pb.InternalComputeExtraInfo{Info: "test"},
	}

	compute_info := pb.InternalComputeConfigInfo{
		OperationType: pb.OperationType_CREATE,
		InternalComputeConfiguration: computeConfig
	}


	resp, err := client.ComputeHandler(ctx, &compute_info)
	if err != nil {
		t.Fatalf("Compute Handler failed: %v", err)
	}
	resp, err = client.TestHandler(ctx, &pb.InternalComputeConfigInfo{})
	if err != nil {
		t.Fatalf("Test Handler failed: %v", err)
	}
	log.Printf("Response: %+v", resp)
}
