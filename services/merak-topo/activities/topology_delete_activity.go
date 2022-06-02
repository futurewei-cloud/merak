package activities

import (
	"context"

	"go.temporal.io/sdk/activity"
)

func TopologyDelete(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)
	return name + "!", nil
}
