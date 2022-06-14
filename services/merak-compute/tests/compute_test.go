package test

import (
	"context"
	"log"
	"net"
	"testing"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-compute/grpc/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewMerakComputeServiceClient(conn)

	pod0 := pb.InternalVMPod{
		OperationType: pb.OperationType_CREATE,
		PodIp:         "10.0.0.2",
		Subnets:       []string{"subnet0", "subnet1"},
		NumOfVm:       3,
	}

	pod1 := pb.InternalVMPod{
		OperationType: pb.OperationType_CREATE,
		PodIp:         "10.0.0.3",
		Subnets:       []string{"subnet0", "subnet1"},
		NumOfVm:       3,
	}
	pod2 := pb.InternalVMPod{
		OperationType: pb.OperationType_CREATE,
		PodIp:         "10.0.0.4",
		Subnets:       []string{"subnet0", "subnet1"},
		NumOfVm:       3,
	}

	subnets := pb.InternalSubnetInfo{
		SubnetId:   "1",
		SubnetCidr: "10.0.0.0/16",
		SubnetGw:   "10.0.0.1",
		NumberVms:  2,
	}
	vpc := pb.InternalVpcInfo{
		VpcId:   "1",
		Subnets: []*pb.InternalSubnetInfo{&subnets},
	}
	deploy := pb.InternalVMDeployInfo{
		OperationType: pb.OperationType_CREATE,
		DeployType:    pb.VMDeployType_UNIFORM,
		Vpcs:          []*pb.InternalVpcInfo{&vpc},
		Secgroups:     []string{"test1", "test2"},
		Scheduler:     pb.VMScheduleType_SEQUENTIAL,
		DeployMethod:  []*pb.InternalVMPod{&pod0, &pod1, &pod2},
	}

	service := pb.InternalServiceInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "2",
		Name:          "test",
		Cmd:           "create",
		Url:           "merak.com",
		Parameters:    []string{"test1", "test2"},
		ReturnCode:    []uint32{0},
		ReturnString:  []string{"success"},
		WhenToRun:     "now",
		WhereToRun:    "here",
	}
	pod := pb.InternalComputeInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "1",
		Name:          "test",
		Ip:            "10.0.0.5",
		Mac:           "aa:bb:cc:dd:ee",
		Veth:          "test",
	}
	computeConfig := pb.InternalComputeConfiguration{
		FormatVersion:   1,
		RevisionNumber:  1,
		RequestId:       "test",
		ComputeConfigId: "test",
		MessageType:     pb.MessageType_FULL,
		Pods:            []*pb.InternalComputeInfo{&pod},
		VmDeploy:        &deploy,
		Services:        []*pb.InternalServiceInfo{&service},
		ExtraInfo:       &pb.InternalComputeExtraInfo{Info: "test"},
	}

	compute_info := pb.InternalComputeConfigInfo{
		OperationType: pb.OperationType_CREATE,
		Config:        &computeConfig,
	}

	resp, err := client.ComputeHandler(ctx, &compute_info)
	if err != nil {
		t.Fatalf("Compute Handler failed: %v", err)
	}

	log.Printf("Response: %+v", resp)
}
