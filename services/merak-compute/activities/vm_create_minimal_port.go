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

	agent_pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	commonPB "github.com/futurewei-cloud/merak/api/proto/v1/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
)

// Creates a VM given by the vmID
func VmCreateMinimalPort(ctx context.Context, vmID string, podIP string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("VmCreateMinimalPort: Starting create activity for VM " + vmID)

	client := common.ClientMapGRPC[podIP]
	logger.Info("VmCreateMinimalPort: Sending to agent at " + podIP)
	port := agent_pb.InternalPortConfig{
		OperationType: commonPB.OperationType_PRECREATE,
		Id:            vmID,
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
		logger.Error("VmCreateMinimalPort: Failed to create vm on " + podIP + "Reason: " + resp.GetReturnMessage() + "\n")
		if err := common.RedisClient.HSet(
			ctx,
			vmID,
			"status",
			"5",
		).Err(); err != nil {
			logger.Info("VmCreateMinimalPort: Failed to add vm response to DB!")
			return err
		}
		return err
	}

	// Update DB with device information
	if resp.ReturnCode == commonPB.ReturnCode_OK {
		ip := resp.Port.GetIp()
		mac := resp.Port.GetMac()
		deviceID := resp.Port.GetDeviceid()
		remoteID := resp.Port.GetRemoteid()
		if err := common.RedisClient.HSet(
			ctx,
			vmID,
			"ip",
			ip,
			"mac",
			mac,
			"deviceID",
			deviceID,
			"remoteID",
			remoteID,
			"status",
			"5",
		).Err(); err != nil {
			logger.Info("VmCreateMinimalPort: Failed to add vm response to DB!")
			return err
		}
	}
	logger.Info("VmCreateMinimalPort: Response from agent at address " + podIP + ": " + resp.GetReturnMessage())
	return nil
}
