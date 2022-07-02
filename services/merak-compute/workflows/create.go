package create

import (
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func Create(ctx workflow.Context) (pb.ReturnMessage, error) {
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 1,
		MaximumInterval:    1,
		MaximumAttempts:    1,
	}
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Second,
		RetryPolicy:         retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	//logger = log.With(logger)
	var result pb.ReturnMessage
	logger.Info("VmCreate starting workflow.")
	err := workflow.ExecuteActivity(ctx, activities.VmCreate).Get(ctx, &result)
	if err != nil {
		logger.Error("VmCreate failed! %s\n", err)
		return pb.ReturnMessage{
			ReturnCode:    result.GetReturnCode(),
			ReturnMessage: result.GetReturnMessage(),
		}, err
	}
	logger.Info("VmCreate workflow completed.%s\n")
	return pb.ReturnMessage{
		ReturnCode:    result.GetReturnCode(),
		ReturnMessage: result.GetReturnMessage(),
	}, nil
}
