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

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
)

func VmInfo(ctx context.Context) (*pb.ReturnComputeMessage, error) {
	logger := activity.GetLogger(ctx)
	ids := common.RedisClient.SMembers(ctx, constants.COMPUTE_REDIS_VM_SET)
	if ids.Err() != nil {
		logger.Error("Unable get VM IDs from redis", ids.Err())

		return &pb.ReturnComputeMessage{
			ReturnCode:    common_pb.ReturnCode_FAILED,
			ReturnMessage: "Unable get node IDs from redis",
		}, ids.Err()
	}
	vms := []*pb.InternalVMInfo{}
	logger.Info("Success in getting VM IDs! " + ids.String())
	for _, vmID := range ids.Val() {
		logger.Info("Found VM ID: " + vmID + "\n")
		vm := pb.InternalVMInfo{
			Name:            common.RedisClient.HGet(ctx, vmID, "name").Val(),
			VpcId:           common.RedisClient.HGet(ctx, vmID, "vpc").Val(),
			Ip:              common.RedisClient.HGet(ctx, vmID, "ip").Val(),
			SecurityGroupId: common.RedisClient.HGet(ctx, vmID, "sh").Val(),
			SubnetId:        common.RedisClient.HGet(ctx, vmID, "subnetID").Val(),
			DefaultGateway:  common.RedisClient.HGet(ctx, vmID, "gw").Val(),
		}
		status, err := strconv.Atoi(common.RedisClient.HGet(ctx, vmID, "status").Val())
		if err != nil {
			logger.Error("Failed to convert status string to int!", err)
			return &pb.ReturnComputeMessage{
				ReturnCode:    common_pb.ReturnCode_FAILED,
				ReturnMessage: "Failed to convert status string to int!",
				Vms:           vms,
			}, err
		}
		vm.Status = common_pb.Status(status)
		vms = append(vms, &vm)
	}

	return &pb.ReturnComputeMessage{
		ReturnCode:    common_pb.ReturnCode_OK,
		ReturnMessage: "Success!",
		Vms:           vms,
	}, nil
}
