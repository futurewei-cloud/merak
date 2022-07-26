package create

import (
	compute_pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func Create(ctx workflow.Context) (compute_pb.ReturnMessage, error) {
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
	//logger = log.With(logger)
	var result compute_pb.ReturnMessage
	logger.Info("VmCreate starting workflow.")
	err := workflow.ExecuteActivity(ctx, activities.VmCreate).Get(ctx, &result)
	if err != nil {
		logger.Error("VmCreate failed! %s\n", err)
		return compute_pb.ReturnMessage{
			ReturnCode:    result.GetReturnCode(),
			ReturnMessage: result.GetReturnMessage(),
			ReturnVms:     result.GetReturnVms(),
		}, err
	}
	logger.Info("VmCreate workflow completed.%s\n")
	return compute_pb.ReturnMessage{
		ReturnCode:    result.GetReturnCode(),
		ReturnMessage: result.GetReturnMessage(),
		ReturnVms:     result.GetReturnVms(),
	}, nil
}
