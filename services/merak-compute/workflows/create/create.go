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
package create

import (
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"github.com/go-redis/redis/v8"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func Create(ctx workflow.Context, vms redis.StringSliceCmd) (pb.ReturnComputeMessage, error) {
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_ACTIVITY_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_ACTIVITY_BACKOFF,
		MaximumInterval:    common.TEMPORAL_ACTIVITY_MAX_INTERVAL,
		MaximumAttempts:    common.TEMPORAL_ACTIVITY_MAX_ATTEMPT,
	}
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: common.TEMPORAL_ACTIVITY_TIMEOUT,
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	var vm_status []*pb.InternalVMInfo
	logger.Info("VmCreate starting workflow.")
	for _, vm := range vms.Val() {
		var vm_result pb.ReturnComputeMessage
		err := workflow.ExecuteActivity(ctx, activities.VmCreate, vm).Get(ctx, &vm_result)
		if err != nil {
			logger.Error("VmCreate failed! %s\n", err)
			return pb.ReturnComputeMessage{
				ReturnCode:    vm_result.GetReturnCode(),
				ReturnMessage: vm_result.GetReturnMessage(),
				Vms:           vm_result.GetVms(),
			}, err
		}
		vm_status = append(vm_status, vm_result.GetVms()...)
	}
	logger.Info("VmCreate workflow completed.\n")
	return pb.ReturnComputeMessage{
		ReturnCode:    common_pb.ReturnCode_OK,
		ReturnMessage: "Success!",
		Vms:           vm_status,
	}, nil
}
