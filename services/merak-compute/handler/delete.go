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

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/delete"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

func caseDelete(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnComputeMessage, error) {

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_WF_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_WF_BACKOFF,
		MaximumInterval:    common.TEMPORAL_WF_MAX_INTERVAL,
		MaximumAttempts:    common.TEMPORAL_WF_MAX_ATTEMPT,
	}

	log.Println("Operation Delete")
	pod_list := RedisClient.SMembers(
		ctx,
		constants.COMPUTE_REDIS_NODE_IP_SET,
	)
	if pod_list.Err() != nil {
		log.Println("Unable get VM IDs from redis", pod_list.Err())

		return &pb.ReturnComputeMessage{
			ReturnCode:    common_pb.ReturnCode_FAILED,
			ReturnMessage: "Unable get node IDs from redis",
		}, pod_list.Err()
	}
	// Get list of all vms in pod
	for n, pod_id := range pod_list.Val() {
		vms := RedisClient.LRange(ctx, "l"+pod_id, 0, -1)
		if vms.Err() != nil {
			log.Println("Unable get node vmIDsList from redis", vms.Err())
			return &pb.ReturnComputeMessage{
				ReturnCode:    common_pb.ReturnCode_FAILED,
				ReturnMessage: "Unable get node vmIDsList from redis",
			}, vms.Err()
		}
		for _, vm_id := range vms.Val() {
			if err := RedisClient.HSet(
				ctx,
				vm_id,
				"status", "3",
			).Err(); err != nil {
				return &pb.ReturnComputeMessage{
					ReturnMessage: "Unable to set VM status to deleting in DB Hash Map",
					ReturnCode:    common_pb.ReturnCode_FAILED,
				}, err
			}
		}
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_DELETE_WORKFLOW_ID + strconv.Itoa(n),
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Executing VM Delete Workflow!")
		we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, delete.Delete, vms.Val())
		if err != nil {
			return &pb.ReturnComputeMessage{
				ReturnMessage: "Unable to execute delete workflow",
				ReturnCode:    common_pb.ReturnCode_FAILED,
			}, err
		}
		log.Println("Started Delete workflow WorkflowID "+we.GetID()+" RunID ", we.GetRunID())
	}

	log.Println("Deleting all VMs from DB")
	RedisClient.FlushAll(ctx)

	return &pb.ReturnComputeMessage{
		ReturnMessage: "Successfully started all delete workflows!",
		ReturnCode:    common_pb.ReturnCode_OK,
	}, nil
}
