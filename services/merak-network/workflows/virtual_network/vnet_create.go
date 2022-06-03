package virtualnetwork

import (
	"time"

	"github.com/futurewei-cloud/merak/services/merak-network/activities"
	"go.temporal.io/sdk/workflow"
)

func Create(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Network Create workflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, activities.VnetCreate, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("Network Create workflow completed.", "result", result)

	return result, nil
}
