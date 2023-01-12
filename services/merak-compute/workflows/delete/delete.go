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
package delete

import (
	"context"
	"strings"

	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	merakwf "github.com/futurewei-cloud/merak/services/merak-compute/workflows"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func Delete(ctx workflow.Context, vms []string, podID string) (err error) {
	var errors error
	defer merakwf.MerakMetrics.GetMetrics(&errors)()
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_ACTIVITY_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_ACTIVITY_BACKOFF,
		MaximumInterval:    common.TEMPORAL_ACTIVITY_MAX_INTERVAL,
		MaximumAttempts:    common.TEMPORAL_ACTIVITY_MAX_ATTEMPT,
	}
	lao := workflow.LocalActivityOptions{
		StartToCloseTimeout: common.TEMPORAL_ACTIVITY_TIMEOUT,
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithLocalActivityOptions(ctx, lao)
	logger := workflow.GetLogger(ctx)

	var futures []workflow.Future
	for _, vm := range vms {
		future := workflow.ExecuteLocalActivity(ctx, activities.VmDelete, vm)
		futures = append(futures, future)
	}
	logger.Info("Started VmDelete workflows for vms" + strings.Join(vms, " "))
	var vmID string
	wfContext := context.Background()
	for _, future := range futures {
		err = future.Get(ctx, &vmID)
		if err != nil {
			logger.Error("Failed to delete VM ID " + vmID)
			return err
		}

		logger.Info("Deleted VM ID " + vmID)

		// Delete Single VM from DB
		common.RedisClient.LRem(wfContext, "l"+podID, 1, vmID)
	}
	// All VMs on pod have been deleted.
	// Delete all Pod/VM associations from DB
	common.RedisClient.HDel(wfContext, podID)
	common.RedisClient.SRem(wfContext, constants.COMPUTE_REDIS_NODE_IP_SET, podID)
	common.RedisClient.Del(wfContext, "l"+podID)
	logger.Info("All VMs for pod " + podID + " deleted!")
	return nil

}
