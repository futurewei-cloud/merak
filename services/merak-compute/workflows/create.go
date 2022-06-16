package create

import (
	"time"

	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"go.temporal.io/sdk/workflow"
)

func Create(ctx workflow.Context) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	var result string
	logger.Info("VmCreate starting workflow.")
	err := workflow.ExecuteActivity(ctx, activities.VmCreate).Get(ctx, &result)
	if err != nil {
		logger.Error("VmCreate failed!", "Error", err)
		return "FAILED", err
	}
	logger.Info("VmCreate workflow completed.", "result", result)
	return result, nil
}
