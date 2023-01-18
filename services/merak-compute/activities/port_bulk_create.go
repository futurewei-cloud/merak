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
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
)

// Calls ovsdb bulk port add
func PortBulkCreate(ctx context.Context, vms []string, podIP string) error {
	logger := activity.GetLogger(ctx)

	client := common.ClientMapGRPC[podIP]
	ports := agent_pb.BulkPorts{}
	logger.Info("PortBulkCreate: Sending to agent at " + podIP)
	tapNames := []string{}
	for _, vmID := range vms {
		tapName := common.RedisClient.HGet(ctx, vmID, "deviceID").Val()
		tapNames = append(tapNames, tapName)
	}
	ports.Tapnames = tapNames

	_, err := client.BulkPortAdd(ctx, &ports)
	if err != nil {
		logger.Info("PortBulkCreate: Failed to BulkAddPort", err)
		return err
	}

	return nil
}
