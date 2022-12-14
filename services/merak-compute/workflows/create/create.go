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
	"strconv"
	"strings"

	agent_pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Create(ctx workflow.Context, vms []string, podIP string) (err error) {
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

	// Workflow is on a per pod basis
	// Each "VMCreate" workflow populates a table mapping podIP -> gRPC client
	var agent_address strings.Builder
	agent_address.WriteString(podIP)
	logger.Info("Connecting to pod at: " + podIP)
	agent_address.WriteString(":")
	agent_address.WriteString(strconv.Itoa(constants.AGENT_GRPC_SERVER_PORT))
	conn, err := grpc.Dial(agent_address.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Info("Failed to dial gRPC server address: "+agent_address.String(), err)
		return err
	}
	client := agent_pb.NewMerakAgentServiceClient(conn)
	common.ClientMapGRPC[podIP] = client

	var futures []workflow.Future
	for _, vm := range vms {
		future := workflow.ExecuteLocalActivity(ctx, activities.VmCreate, vm, podIP)
		logger.Info("VmCreate activity started for vm_id " + vm)
		futures = append(futures, future)
	}
	logger.Info("Started VmCreate workflows for vms" + strings.Join(vms, " "))

	for _, future := range futures {
		err = future.Get(ctx, nil)
		logger.Info("Activity completed!")
		if err != nil {
			return nil
		}
	}
	logger.Info("All activities completed")
	defer conn.Close()
	return nil
}
