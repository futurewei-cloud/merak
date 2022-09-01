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
package info

import (
	compute_pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func Info(ctx workflow.Context) (compute_pb.ReturnComputeMessage, error) {
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

	var result compute_pb.ReturnComputeMessage
	logger.Info("VmInfo starting workflow.")
	err := workflow.ExecuteActivity(ctx, activities.VmInfo).Get(ctx, &result)
	if err != nil {
		logger.Error("VmInfo failed! %s\n", err)
		return compute_pb.ReturnComputeMessage{
			ReturnCode:    result.GetReturnCode(),
			ReturnMessage: result.GetReturnMessage(),
			Vms:           result.GetVms(),
		}, err
	}
	logger.Info("VmInfo workflow completed!")
	return compute_pb.ReturnComputeMessage{
		ReturnCode:    result.GetReturnCode(),
		ReturnMessage: result.GetReturnMessage(),
		Vms:           result.GetVms(),
	}, nil
}
