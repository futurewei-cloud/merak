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

	commonPB "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/ntest"
	"github.com/futurewei-cloud/merak/services/merak-ntest/common"
	"github.com/futurewei-cloud/merak/services/merak-ntest/workflows/create"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

func caseCreate(ctx context.Context, in *pb.InternalTestConfiguration) (*pb.ReturnTestMessage, error) {

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_WF_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_WF_BACKOFF,
		MaximumInterval:    common.TEMPORAL_WF_MAX_INTERVAL,
		MaximumAttempts:    common.TEMPORAL_WF_MAX_ATTEMPT,
	}

	log.Println("Operation Create")

	// Execute VM creation on a per pod basis
	// Send a list of VMs to the Workflow
	workflowOptions = client.StartWorkflowOptions{
		ID:          common.NTEST_CREATE_WORKFLOW_ID,
		TaskQueue:   common.NTEST_TASK_QUEUE,
		RetryPolicy: retrypolicy,
	}
	log.Println("Executing Ntest Create Workflow with")
	we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, create.Create, in)
	if err != nil {
		return &pb.ReturnTestMessage{
			ReturnMessage: "Unable to execute Ntest create workflow",
		}, err
	}
	log.Println("Started Create workflow WorkflowID "+we.GetID()+" RunID ", we.GetRunID())

	return &pb.ReturnTestMessage{
		ReturnMessage: "Successfully started all create workflows!",
		ReturnCode:    commonPB.ReturnCode_OK,
	}, nil
}
