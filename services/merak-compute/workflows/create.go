package create

import (
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"go.temporal.io/sdk/workflow"
)

func Create(ctx workflow.Context) (pb.ReturnMessage, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Second,
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
			ReturnCode:    pb.ReturnCode_FAILED,
			ReturnMessage: "VmCreate failed!",
		}, err
	}
	logger.Info("VmCreate workflow completed.%s\n")
	return pb.ReturnMessage{
		ReturnCode:    result.ReturnCode,
		ReturnMessage: result.ReturnMessage,
	}, nil
}
