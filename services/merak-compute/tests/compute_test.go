package test

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	"github.com/futurewei-cloud/merak/services/merak-compute/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterMerakComputeServiceServer(s, &handler.Server{})
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

	fmt.Println("Pod IP Address: ")
	var ip string
	fmt.Scanln(&ip)

	pod0 := pb.InternalVMPod{
		OperationType: common_pb.OperationType_CREATE,
		PodIp:         ip,
		Subnets:       []string{"subnet0", "subnet1"},
		NumOfVm:       10,
	}

	pod1 := pb.InternalVMPod{
		OperationType: common_pb.OperationType_CREATE,
		PodIp:         ip,
		Subnets:       []string{"subnet0", "subnet1"},
		NumOfVm:       10,
	}

	subnets := common_pb.InternalSubnetInfo{
		SubnetId:   "1",
		SubnetCidr: "10.0.0.0/16",
		SubnetGw:   "10.0.0.1",
		NumberVms:  10,
	}
	vpc := common_pb.InternalVpcInfo{
		VpcId:     "9192a4d4-ffff-4ece-b3f0-8d36e3d88001",
		Subnets:   []*common_pb.InternalSubnetInfo{&subnets},
		ProjectId: "123456789",
		TenantId:  "123456789",
	}
	deploy := pb.InternalVMDeployInfo{
		OperationType: common_pb.OperationType_CREATE,
		DeployType:    pb.VMDeployType_UNIFORM,
		Vpcs:          []*common_pb.InternalVpcInfo{&vpc},
		Secgroups:     []string{"test1", "test2"},
		Scheduler:     pb.VMScheduleType_SEQUENTIAL,
		DeployMethod:  []*pb.InternalVMPod{&pod0, &pod1},
	}

	service := common_pb.InternalServiceInfo{
		OperationType: common_pb.OperationType_CREATE,
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
	pod := common_pb.InternalComputeInfo{
		OperationType: common_pb.OperationType_CREATE,
		Id:            "1",
		Name:          "test",
		Ip:            ip,
		Mac:           "aa:bb:cc:dd:ee",
		Veth:          "test",
	}
	computeConfig := pb.InternalComputeConfiguration{
		FormatVersion:   1,
		RevisionNumber:  1,
		RequestId:       "test",
		ComputeConfigId: "test",
		MessageType:     common_pb.MessageType_FULL,
		Pods:            []*common_pb.InternalComputeInfo{&pod},
		VmDeploy:        &deploy,
		Services:        []*common_pb.InternalServiceInfo{&service},
		ExtraInfo:       &pb.InternalComputeExtraInfo{Info: "test"},
	}

	compute_info := pb.InternalComputeConfigInfo{
		OperationType: common_pb.OperationType_CREATE,
		Config:        &computeConfig,
	}

	resp, err := client.ComputeHandler(ctx, &compute_info)
	if err != nil {
		t.Fatalf("Compute Handler failed: %v", err)
	}

	log.Printf("Response: %+v", resp)
}
