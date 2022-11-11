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
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/create"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
)

func caseCreate(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnComputeMessage, error) {

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_WF_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_WF_BACKOFF,
		MaximumInterval:    common.TEMPORAL_WF_MAX_INTERVAL,
		MaximumAttempts:    common.TEMPORAL_WF_MAX_ATTEMPT,
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Infow("Failed to fetch URL")

	log.Println("Operation Create")
	returnVMs := []*pb.InternalVMInfo{}
	// Add pods to DB
	count := 0
	for n, pod := range in.Config.Pods {
		if err := RedisClient.HSet(
			ctx,
			pod.Id,
			"name", pod.Name,
			"ip", pod.ContainerIp,
			"mac", pod.Mac,
			"veth", pod.Veth,
		).Err(); err != nil {
			return &pb.ReturnComputeMessage{
				ReturnMessage: "Unable add pod to DB Hash Map",
				ReturnCode:    commonPB.ReturnCode_FAILED,
			}, err
		}
		log.Println("Added pod " + pod.Name + " at address " + pod.ContainerIp)
		if err := RedisClient.SAdd(
			ctx,
			constants.COMPUTE_REDIS_NODE_IP_SET,
			pod.Id,
		).Err(); err != nil {
			return &pb.ReturnComputeMessage{
				ReturnMessage: "Unable to add pod to DB Hash Set",
				ReturnCode:    commonPB.ReturnCode_FAILED,
			}, err
		}

		for i, vpc := range in.Config.VmDeploy.Vpcs {
			for j, subnet := range vpc.Subnets {
				for k := 0; k < int(subnet.NumberVms); k++ {
					vmID := pod.Id + strconv.Itoa(i) + strconv.Itoa(j) + strconv.Itoa(k)
					suffix := strconv.Itoa(i) + strconv.Itoa(j) + strconv.Itoa(k)
					if err := RedisClient.SAdd(
						ctx,
						constants.COMPUTE_REDIS_VM_SET,
						vmID,
					).Err(); err != nil {
						return &pb.ReturnComputeMessage{
							ReturnMessage: "Unable to VM to DB Hash Set",
							ReturnCode:    commonPB.ReturnCode_FAILED,
						}, err
					}
					if err := RedisClient.HSet(
						ctx,
						vmID,
						"id", vmID,
						"name", "v"+suffix,
						"vpc", vpc.VpcId,
						"tenantID", vpc.TenantId,
						"projectID", vpc.ProjectId,
						"subnetID", subnet.SubnetId,
						"cidr", subnet.SubnetCidr,
						"gw", subnet.SubnetGw,
						"sg", in.Config.VmDeploy.Secgroups[0],
						"hostIP", pod.ContainerIp,
						"hostmac", pod.Mac,
						"hostname", pod.Name,
						"status", "1",
					).Err(); err != nil {
						return &pb.ReturnComputeMessage{
							ReturnMessage: "Unable add VM to DB Hash Map",
							ReturnCode:    commonPB.ReturnCode_FAILED,
						}, err
					}

					// Store VM to Pod list
					if err := RedisClient.LPush(ctx, "l"+pod.Id, vmID).Err(); err != nil {
						log.Println("Failed to add pod -> vm mapping " + vmID)
						return &pb.ReturnComputeMessage{
							ReturnMessage: "Unable add VM to pod list",
							ReturnCode:    commonPB.ReturnCode_FAILED,
							Vms:           returnVMs,
						}, err
					}
				}
			}
		}

		// Get VM to pod list
		vms := RedisClient.LRange(ctx, "l"+pod.Id, 0, -1)
		if vms.Err() != nil {
			log.Println("Unable get node vmIDsList from redis", vms.Err())
			return &pb.ReturnComputeMessage{
				ReturnCode:    commonPB.ReturnCode_FAILED,
				ReturnMessage: "Unable get node vmIDsList from redis",
				Vms:           returnVMs,
			}, vms.Err()
		}

		// Execute VM creation on a per pod basis
		// Send a list of VMs to the Workflow
		workflowOptions = client.StartWorkflowOptions{
			ID:                       common.VM_CREATE_WORKFLOW_ID + strconv.Itoa(n),
			TaskQueue:                common.VM_TASK_QUEUE,
			RetryPolicy:              retrypolicy,
			WorkflowExecutionTimeout: common.TEMPORAL_WF_EXEC_TIMEOUT,
			WorkflowRunTimeout:       common.TEMPORAL_WF_RUN_TIMEOUT,
			WorkflowTaskTimeout:      common.TEMPORAL_WF_TASK_TIMEOUT,
		}
		num_vms := strconv.Itoa(len(vms.Val()))
		count += len(vms.Val())
		log.Println("Executing VM Create Workflow with VMs " + num_vms + " on pod at " + pod.ContainerIp)
		_, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, create.Create, vms.Val(), pod.ContainerIp)
		if err != nil {
			return &pb.ReturnComputeMessage{
				ReturnMessage: "Unable to execute create workflow",
				ReturnCode:    commonPB.ReturnCode_FAILED,
				Vms:           returnVMs,
			}, err
		}
	}

	returnVM := pb.InternalVMInfo{
		Id: "Started deployment for " + strconv.Itoa(count) + " VMs",
	}
	returnVMs = append(returnVMs, &returnVM)
	log.Println("Started deployment for " + strconv.Itoa(count) + " VMs")
	return &pb.ReturnComputeMessage{
		ReturnMessage: "Successfully started all create workflows!",
		ReturnCode:    commonPB.ReturnCode_OK,
		Vms:           returnVMs,
	}, nil
}
