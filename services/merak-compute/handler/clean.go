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
	"strconv"

	commonPB "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/clean"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

func caseClean(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnComputeMessage, error) {

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_WF_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_WF_BACKOFF,
		MaximumInterval:    common.TEMPORAL_WF_MAX_INTERVAL,
		MaximumAttempts:    common.TEMPORAL_WF_MAX_ATTEMPT,
	}
	podList := RedisClient.SMembers(
		ctx,
		constants.COMPUTE_REDIS_NODE_IP_SET,
	)
	if podList.Err() != nil {
		log.Println("Unable get Pod IDs from redis", podList.Err())

		return &pb.ReturnComputeMessage{
			ReturnCode:    commonPB.ReturnCode_FAILED,
			ReturnMessage: "Unable get node IDs from redis",
		}, podList.Err()
	}
	// Get list of all pod IPs
	var podIPList []string
	for _, podID := range podList.Val() {
		podIP := RedisClient.HGet(ctx, podID, "ip").Val()
		podIPList = append(podIPList, podIP)
	}

	// Cancel any running workflows
	numVMs := RedisClient.SCard(ctx, constants.COMPUTE_REDIS_NODE_IP_SET).Val()
	var i int64 = 0
	for i < numVMs {
		wfID := common.VM_CREATE_WORKFLOW_ID + strconv.FormatInt(i, 10)
		err := TemporalClient.CancelWorkflow(context.Background(), wfID, "")
		if err != nil {
			log.Println("Unable to cancel Workflow Execution ID "+wfID, err)
		}
		i++
	}
	workflowOptions = client.StartWorkflowOptions{
		ID:          common.VM_CLEAN_WORKFLOW_ID,
		TaskQueue:   common.VM_TASK_QUEUE,
		RetryPolicy: retrypolicy,
	}
	log.Println("Executing VM Clean Workflow!")
	we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, clean.Clean, podIPList)
	if err != nil {
		return &pb.ReturnComputeMessage{
			ReturnMessage: "Unable to execute Clean workflow",
			ReturnCode:    commonPB.ReturnCode_FAILED,
		}, err
	}
	log.Println("Started Clean workflow WorkflowID "+we.GetID()+" RunID ", we.GetRunID())
	RedisClient.FlushAll(ctx)
	return &pb.ReturnComputeMessage{
		ReturnMessage: "Successfully started clean procedure on all pods!",
		ReturnCode:    commonPB.ReturnCode_OK,
	}, nil
}
