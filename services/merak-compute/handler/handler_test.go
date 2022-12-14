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
package handler

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v9"

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
)

func TestHandler(t *testing.T) {

	ip := "10.0.0.2"
	hostname := "host0"
	ctx := context.Background()
	pod0 := pb.InternalVMPod{
		OperationType: common_pb.OperationType_CREATE,
		PodIp:         ip,
		Subnets:       []string{"subnet1"},
		NumOfVm:       1,
	}

	subnets := common_pb.InternalSubnetInfo{
		SubnetId:   "8182a4d4-ffff-4ece-b3f0-8d36e3d88000",
		SubnetCidr: "10.0.1.0/24",
		SubnetGw:   "10.0.1.1",
		NumberVms:  5000,
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

	server, _ := miniredis.Run()

	RedisClient = *redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	caseCreate(ctx, &compute_info)
	caseDelete(ctx, &compute_info)
	caseInfo(ctx, &compute_info)
}
