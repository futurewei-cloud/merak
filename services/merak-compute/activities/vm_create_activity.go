package activities

import (
	"context"
	"fmt"

	"github.com/futurewei-cloud/merak/services/merak-compute/activities/common"
	"go.temporal.io/sdk/activity"
)

func VmCreate(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	ips := common.RedisClient.SMembers(ctx, "NodIPsSet")
	err := ips.Err()
	if err != nil {
		logger.Error("Unable get node IPs from redis", err)
		return "Failed!", err
	}
	logger.Info("Success in storing Node IPs!", "NodeIps", ips.Val())
	fmt.Println("HI FROM VM CREATE")
	logger.Info("HI FROM LOGGER")

	return "success!", nil
}
