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
package test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
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

	var ip string = "10.244.0.91"
	var hostname string = "merak-agent-55847876d9-rp5n8"

	pod0 := pb.InternalVMPod{
		OperationType: common_pb.OperationType_CREATE,
		PodIp:         ip,
		Subnets:       []string{"subnet1"},
		NumOfVm:       10,
	}

	subnets := common_pb.InternalSubnetInfo{
		SubnetId:   "8182a4d4-ffff-4ece-b3f0-8d36e3d88000",
		SubnetCidr: "10.0.1.0/24",
		SubnetGw:   "10.0.1.1",
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
		Secgroups:     []string{"3dda2801-d675-4688-a63f-dcda8d111111"},
		Scheduler:     pb.VMScheduleType_SEQUENTIAL,
		DeployMethod:  []*pb.InternalVMPod{&pod0},
	}

	service := common_pb.InternalServiceInfo{
		OperationType: common_pb.OperationType_CREATE,
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
	pod := common_pb.InternalComputeInfo{
		OperationType: common_pb.OperationType_CREATE,
		Id:            "1",
		Name:          hostname,
		DatapathIp:    ip,
		ContainerIp:   ip,
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

	// Test Create
	resp, err := client.ComputeHandler(ctx, &compute_info)
	if err != nil {
		t.Fatalf("Compute Handler Create failed: %v", err)
	}
	t.Log("Response: ", resp.ReturnMessage)

	// Test Info
	compute_info.OperationType = common_pb.OperationType_INFO
	resp, err = client.ComputeHandler(ctx, &compute_info)
	if err != nil {
		t.Fatalf("Compute Handler Info failed: %v", err)
	}
	t.Log("Response: ", resp.ReturnMessage)

	// Test Delete
	compute_info.OperationType = common_pb.OperationType_DELETE
	resp, err = client.ComputeHandler(ctx, &compute_info)
	if err != nil {
		t.Fatalf("Compute Handler Delete failed: %v", err)
	}
	t.Log("Response: ", resp)

	defer conn.Close()
}
