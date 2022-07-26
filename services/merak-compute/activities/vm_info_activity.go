package activities

import (
	"context"
	"strconv"

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	compute_pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
)

func VmInfo(ctx context.Context) (*compute_pb.ReturnMessage, error) {
	logger := activity.GetLogger(ctx)
	ids := common.RedisClient.SMembers(ctx, constants.COMPUTE_REDIS_VM_SET)
	if ids.Err() != nil {
		logger.Error("Unable get VM IDs from redis", ids.Err())

		return &compute_pb.ReturnMessage{
			ReturnCode:    common_pb.ReturnCode_FAILED,
			ReturnMessage: "Unable get node IDs from redis",
		}, ids.Err()
	}
	vms := []*compute_pb.InternalVMInfo{}
	logger.Info("Success in getting VM IDs! " + ids.String())
	for _, vmID := range ids.Val() {
		vmIDsList := common.RedisClient.LRange(ctx, vmID, 0, -1)
		if vmIDsList.Err() != nil {
			logger.Error("Unable get node vmIDsList from redis", vmIDsList.Err())
			return &compute_pb.ReturnMessage{
				ReturnCode:    common_pb.ReturnCode_FAILED,
				ReturnMessage: "Unable get node vmIDsList from redis",
			}, vmIDsList.Err()
		}
		logger.Info("VM Ids " + vmIDsList.String() + "\n")
		vm := compute_pb.InternalVMInfo{
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
			return &compute_pb.ReturnMessage{
				ReturnCode:    common_pb.ReturnCode_FAILED,
				ReturnMessage: "Failed to convert status string to int!",
				ReturnVms:     vms,
			}, err
		}
		vm.Status = common_pb.Status(status)
		vms = append(vms, &vm)
	}

	return &compute_pb.ReturnMessage{
		ReturnCode:    common_pb.ReturnCode_OK,
		ReturnMessage: "Success!",
		ReturnVms:     vms,
	}, nil
}
