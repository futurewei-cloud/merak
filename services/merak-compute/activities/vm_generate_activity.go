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

	commonPB "github.com/futurewei-cloud/merak/api/proto/v1/common"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
)

// Deletes a VM given by the vmID
func VmGenerate(ctx context.Context,
	pod *commonPB.InternalComputeInfo,
	vpc *commonPB.InternalVpcInfo,
	subnet *commonPB.InternalSubnetInfo,
	sg string,
	i int, j int, k int) error {
	logger := activity.GetLogger(ctx)
	suffix := strconv.Itoa(i) + strconv.Itoa(j) + strconv.Itoa(k)
	logger.Info("Starting create activity for VM v" + suffix)

	vmID := pod.Id + strconv.Itoa(i) + strconv.Itoa(j) + strconv.Itoa(k)

	if err := common.RedisClient.SAdd(
		ctx,
		constants.COMPUTE_REDIS_VM_SET,
		vmID,
	).Err(); err != nil {
		return err
	}
	if err := common.RedisClient.HSet(
		ctx,
		vmID,
		"id", vmID,
		"name", "v"+suffix,
		"vpc", vpc.VpcId,
		"tenantID", vpc.TenantId,
		"projectID", vpc.ProjectId,
		"subnetID", subnet.SubnetId,
		"cidr", subnet.SubnetCidr,
		"gw", subnet.SubnetGw,
		"sg", sg,
		"hostIP", pod.ContainerIp,
		"hostmac", pod.Mac,
		"hostname", pod.Name,
		"status", "1",
	).Err(); err != nil {
		return err
	}

	// Store VM to Pod list
	logger.Info("Added VM " + vmID + " for vpc " + vpc.VpcId + " for subnet " + subnet.SubnetId + " vm number " + strconv.Itoa(k+1) + " of " + strconv.Itoa(int(subnet.NumberVms)))
	if err := common.RedisClient.LPush(ctx, "l"+pod.Id, vmID).Err(); err != nil {
		logger.Info("Failed to add pod -> vm mapping " + vmID)
		return err
	}
	return nil
}
