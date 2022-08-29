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

package handler

import (
	"context"
	"log"

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/info"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

func caseInfo(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnComputeMessage, error) {
	log.Println("Operation Info")

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_WF_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_WF_BACKOFF,
		MaximumInterval:    common.TEMPORAL_WF_MAX_INTERVAL,
		MaximumAttempts:    common.TEMPORAL_WF_MAX_ATTEMPT,
	}

	workflowOptions = client.StartWorkflowOptions{
		ID:                       common.VM_INFO_WORKFLOW_ID,
		TaskQueue:                common.VM_TASK_QUEUE,
		WorkflowTaskTimeout:      common.TEMPORAL_WF_TASK_TIMEOUT,
		WorkflowExecutionTimeout: common.TEMPORAL_WF_EXEC_TIMEOUT,
		WorkflowRunTimeout:       common.TEMPORAL_WF_RUN_TIMEOUT,
		RetryPolicy:              retrypolicy,
	}
	// Start VM Info Workflow
	var result pb.ReturnComputeMessage
	log.Println("Executing VM Info Workflow!")
	we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, info.Info)
	if err != nil {
		return &pb.ReturnComputeMessage{
			ReturnMessage: "Unable to execute info workflow",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, err
	}
	log.Println("Started workflow WorkflowID "+we.GetID()+" RunID ", we.GetRunID())

	// Sync get results of workflow
	err = we.Get(context.Background(), &result)
	if err != nil {
		return &pb.ReturnComputeMessage{
			ReturnMessage: result.GetReturnMessage(),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Vms:           result.GetVms(),
		}, err
	}
	log.Println("Workflow result:", result.ReturnMessage)
	return &pb.ReturnComputeMessage{
		ReturnMessage: result.GetReturnMessage(),
		ReturnCode:    result.GetReturnCode(),
		Vms:           result.GetVms(),
	}, err
}
