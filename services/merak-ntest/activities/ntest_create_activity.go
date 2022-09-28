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
package activities

import (
	"context"
	"strconv"
	"strings"

	agent_pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	commonPB "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/ntest"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-ntest/common"
	"go.temporal.io/sdk/activity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Creates a VM given by the vmID
func NtestCreate(ctx context.Context, vm *pb.InternalVMTestInfo) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Ntest create activity for VM " + vm.Src.Id)

	var agent_address strings.Builder
	podIP := common.RedisClient.HGet(ctx, vm.Src.Id, "hostIP").Val()
	agent_address.WriteString(podIP)
	logger.Info("Connecting to pod at: " + podIP)
	agent_address.WriteString(":")
	agent_address.WriteString(strconv.Itoa(constants.AGENT_GRPC_SERVER_PORT))
	conn, err := grpc.Dial(agent_address.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Info("Failed to dial gRPC server address: "+agent_address.String(), err)
		return err
	}

	client := agent_pb.NewMerakAgentServiceClient(conn)
	logger.Info("Sending to agent at" + podIP)
	name := common.RedisClient.HGet(ctx, vm.Src.Id, "name").Val()
	destName := common.RedisClient.HGet(ctx, vm.Dest.Id, "name").Val()
	destIP := common.RedisClient.HGet(ctx, vm.Dest.Id, "ip").Val()
	srcIP := common.RedisClient.HGet(ctx, vm.Src.Id, "ip").Val()
	test := agent_pb.InternalTestConfig{
		TestType: pb.TestType_PING,
		Name:     name,
		DestIp:   destIP,
		SrcPort:  vm.Src.Port,
		DestPort: vm.Dest.Port,
	}
	resp, err := client.TestHandler(ctx, &test)
	if err != nil {
		logger.Error("Unable to start vm test on" + podIP + "Reason: " + resp.GetReturnMessage() + "\n")
		if err := common.RedisClient.HSet(
			ctx,
			constants.TEST_PREFIX+vm.Src.Id,
			"status",
			"2",
		).Err(); err != nil {
			logger.Info("Failed to add err test response to DB!")
			return err
		}
		return err
	}

	// Update DB with device information
	if resp.ReturnCode == commonPB.ReturnCode_OK {
		status := resp.Status
		logger.Info("Test Success for " + name + "on pod IP " + podIP)
		if err := common.RedisClient.HSet(
			ctx,
			constants.TEST_PREFIX+vm.Src.Id,
			"id",
			vm.Src.Id,
			"name",
			name,
			"src",
			srcIP,
			"dest",
			destIP,
			"srcPort",
			vm.Src.Port,
			"destPort",
			vm.Dest.Port,
			"destID",
			vm.Dest.Id,
			"destName",
			destName,
			"status",
			strconv.Itoa(int(status.Number())),
		).Err(); err != nil {
			logger.Info("Failed to add Test response to DB!")
			return err
		}
	}
	logger.Info("Response from agent at address " + podIP + ": " + resp.GetReturnMessage())

	defer conn.Close()
	return nil
}
