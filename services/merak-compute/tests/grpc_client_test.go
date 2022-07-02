package test

import (
	"context"
	"log"
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
	ctx := context.Background()
	conn, err := grpc.Dial(compute_address.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial gRPC server address!: %v", err)
	}
	client := pb.NewMerakComputeServiceClient(conn)

	var ip string = ""

	pod0 := pb.InternalVMPod{
		OperationType: pb.OperationType_CREATE,
		PodIp:         ip,
		Subnets:       []string{"subnet1"},
		NumOfVm:       1,
	}

	subnets := pb.InternalSubnetInfo{
		SubnetId:   "8182a4d4-ffff-4ece-b3f0-8d36e3d88000",
		SubnetCidr: "10.0.1.0/24",
		SubnetGw:   "10.0.1.1",
		NumberVms:  1,
	}
	vpc := pb.InternalVpcInfo{
		VpcId:     "9192a4d4-ffff-4ece-b3f0-8d36e3d88001",
		Subnets:   []*pb.InternalSubnetInfo{&subnets},
		ProjectId: "123456789",
		TenantId:  "123456789",
	}
	deploy := pb.InternalVMDeployInfo{
		OperationType: pb.OperationType_CREATE,
		DeployType:    pb.VMDeployType_UNIFORM,
		Vpcs:          []*pb.InternalVpcInfo{&vpc},
		Secgroups:     []string{"3dda2801-d675-4688-a63f-dcda8d111111"},
		Scheduler:     pb.VMScheduleType_SEQUENTIAL,
		DeployMethod:  []*pb.InternalVMPod{&pod0},
	}

	service := pb.InternalServiceInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "2",
		Name:          "test",
		Cmd:           "10.224",
		Url:           "project/123456789/ports",
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
		Ip:            ip,
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

	t.Logf("Response: %+v", resp)
	defer conn.Close()
}
