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
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func VmCreate(ctx context.Context, vmID string) {
	logger := activity.GetLogger(ctx)

	var agent_address strings.Builder
	podIP := common.RedisClient.HGet(ctx, vmID, "hostIP").Val()
	agent_address.WriteString(podIP)
	logger.Info("Connecting to pod at: ", podIP)
	agent_address.WriteString(":")
	agent_address.WriteString(strconv.Itoa(constants.AGENT_GRPC_SERVER_PORT))
	conn, err := grpc.Dial(agent_address.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Info("Failed to dial gRPC server address: "+agent_address.String(), err)
		return
	}
	client := agent_pb.NewMerakAgentServiceClient(conn)

	logger.Info("Sending to agent at" + podIP)
	port := agent_pb.InternalPortConfig{
		OperationType: common_pb.OperationType_CREATE,
		Name:          common.RedisClient.HGet(ctx, vmID, "name").Val(),
		Vpcid:         common.RedisClient.HGet(ctx, vmID, "vpc").Val(),
		Tenantid:      common.RedisClient.HGet(ctx, vmID, "tenantID").Val(),
		Projectid:     common.RedisClient.HGet(ctx, vmID, "projectID").Val(),
		Subnetid:      common.RedisClient.HGet(ctx, vmID, "subnetID").Val(),
		Gw:            common.RedisClient.HGet(ctx, vmID, "gw").Val(),
		Sg:            common.RedisClient.HGet(ctx, vmID, "sg").Val(),
		Cidr:          common.RedisClient.HGet(ctx, vmID, "cidr").Val(),
		Hostname:      common.RedisClient.HGet(ctx, vmID, "hostname").Val(),
	}
	resp, err := client.PortHandler(ctx, &port)
	if err != nil {
		logger.Error("Unable to create vm on" + podIP + "Reason: " + resp.GetReturnMessage() + "\n")
		if err := common.RedisClient.HSet(
			ctx,
			vmID,
			"status",
			"5",
		).Err(); err != nil {
			logger.Info("Failed to add vm response to DB!")
			return
		}
		return
	}

	if resp.ReturnCode == common_pb.ReturnCode_OK {
		ip := resp.Port.GetIp()
		status := resp.Port.GetStatus()
		deviceID := resp.Port.GetDeviceid()
		remoteID := resp.Port.GetRemoteid()
		logger.Info("IP:" + ip + " status:" + status.String() + " deviceID:" + deviceID + " remoteID:" + remoteID)
		if err := common.RedisClient.HSet(
			ctx,
			vmID,
			"ip",
			ip,
			"status",
			strconv.Itoa(int(status.Number())),
			"deviceID",
			deviceID,
			"remoteID",
			remoteID,
		).Err(); err != nil {
			logger.Info("Failed to add vm response to DB!")
			return
		}
	}
	logger.Info("Response from agent at address " + podIP + ": " + resp.GetReturnMessage())

	defer conn.Close()
}
