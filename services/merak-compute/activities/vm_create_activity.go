package activities

import (
	"context"

	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

func VmCreate(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger = log.With(logger)
	ids := common.RedisClient.SMembers(ctx, constants.COMPUTE_REDIS_NODE_IP_SET)
	if ids.Err() != nil {
		logger.Error("Unable get node IDs from redis", ids.Err())
		return "Failed!", ids.Err()
	}
	logger.Info("Success in getting Node IDs!", "NodeIDs", ids.Val())

	for _, podID := range ids.Val() {
		vmIDsList := common.RedisClient.LRange(ctx, podID, 0, -1)
		if vmIDsList.Err() != nil {
			logger.Error("Unable get node vmIDsList from redis", vmIDsList.Err())
			return "Failed!", vmIDsList.Err()
		}
		logger.Info("Pod IDs", ids.String())
		logger.Info("VM Ids", vmIDsList.String())
		for _, vmID := range vmIDsList.Val() {
			logger.Info("Looking up ID ", vmID)
			vm := common.RedisClient.HGetAll(ctx, vmID)
			if vm.Err() != nil {
				logger.Error("Unable get node VM from redis for", "vmID", vmID, vm.Err())
				return "Failed!", vm.Err()
			}
			logger.Info(vm.String())
		}
	}

	return common.TEMPORAL_SUCESS_CODE, nil
}
