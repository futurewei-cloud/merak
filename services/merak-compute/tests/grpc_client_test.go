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
	vmPod := pb.InternalVMPod{
		OperationType: pb.OperationType_CREATE,
		PodIp:         "10.0.0.2",
		NumOfVm:       3,
		Subnets:       []string{"subnet0", "subnet1"},
	}

	pod0 := pb.InternalPodInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "0",
		Name:          "pod_name",
		Ip:            "10.0.0.2",
	}

	pod1 := pb.InternalPodInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "1",
		Name:          "pod_name",
		Ip:            "10.0.0.3",
	}
	pod2 := pb.InternalPodInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "2",
		Name:          "pod_name",
		Ip:            "10.0.0.4",
	}
	vms := pb.InternalVMInfo{
		OperationType: pb.OperationType_CREATE,
		DeployType:    pb.VMDeployType_RANDOM,
		Vpcs:          []string{"vpc0", "vpc1"},
		Subnets:       []string{"subnet0", "subnet1"},
		Secgroups:     []string{"sg0", "sg1"},
		NumPortPerVm:  2,
		DeployMethod:  []*pb.InternalVMPod{&vmPod},
	}
	computeConfig := pb.InternalComputeConfiguration{
		FormatVersion:   1,
		RevisionNumber:  1,
		RequestId:       "test",
		ComputeConfigId: "test",
		MessageType:     pb.MessageType_FULL,
		Pods:            []*pb.InternalPodInfo{&pod0, &pod1, &pod2},
		Vms:             []*pb.InternalVMInfo{&vms},
		Scheduler:       pb.VMScheduleType_SEQUENTIAL,
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
