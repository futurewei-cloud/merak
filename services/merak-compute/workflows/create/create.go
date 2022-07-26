package create

import (
<<<<<<< HEAD:services/merak-compute/workflows/create.go
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
=======
	compute_pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
>>>>>>> 26f1812 (Compute INFO implementation):services/merak-compute/workflows/create/create.go
	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

<<<<<<< HEAD:services/merak-compute/workflows/create.go
func Create(ctx workflow.Context) (pb.ReturnComputeMessage, error) {
=======
func Create(ctx workflow.Context) (compute_pb.ReturnMessage, error) {
>>>>>>> 26f1812 (Compute INFO implementation):services/merak-compute/workflows/create/create.go
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
<<<<<<< HEAD:services/merak-compute/workflows/create.go
	var result pb.ReturnComputeMessage
=======
	var result compute_pb.ReturnMessage
>>>>>>> 26f1812 (Compute INFO implementation):services/merak-compute/workflows/create/create.go
	logger.Info("VmCreate starting workflow.")
	err := workflow.ExecuteActivity(ctx, activities.VmCreate).Get(ctx, &result)
	if err != nil {
		logger.Error("VmCreate failed! %s\n", err)
<<<<<<< HEAD:services/merak-compute/workflows/create.go
		return pb.ReturnComputeMessage{
=======
		return compute_pb.ReturnMessage{
>>>>>>> 26f1812 (Compute INFO implementation):services/merak-compute/workflows/create/create.go
			ReturnCode:    result.GetReturnCode(),
			ReturnMessage: result.GetReturnMessage(),
			Vms:           result.GetVms(),
		}, err
	}
	logger.Info("VmCreate workflow completed.%s\n")
<<<<<<< HEAD:services/merak-compute/workflows/create.go
	return pb.ReturnComputeMessage{
=======
	return compute_pb.ReturnMessage{
>>>>>>> 26f1812 (Compute INFO implementation):services/merak-compute/workflows/create/create.go
		ReturnCode:    result.GetReturnCode(),
		ReturnMessage: result.GetReturnMessage(),
		Vms:           result.GetVms(),
	}, nil
}
