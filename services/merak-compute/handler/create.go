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
	"golang.org/x/sync/errgroup"
)

func caseCreate(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnComputeMessage, error) {

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_WF_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_WF_BACKOFF,
		MaximumInterval:    common.TEMPORAL_WF_MAX_INTERVAL,
		MaximumAttempts:    common.TEMPORAL_WF_MAX_ATTEMPT,
	}
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
		gen := new(errgroup.Group)
		//Assume same number of Subnets per VPC and same number of VMs per subnet for now
		vms := make([]string,
			len(in.Config.VmDeploy.Vpcs)*len(in.Config.VmDeploy.Vpcs[0].Subnets)*int(in.Config.VmDeploy.Vpcs[0].Subnets[0].NumberVms))
		for i, vpc := range in.Config.VmDeploy.Vpcs {
			for j, subnet := range vpc.Subnets {
				for k := 0; k < int(subnet.NumberVms); k++ {
					func(i, j, k int,
						subnet *commonPB.InternalSubnetInfo,
						vpc *commonPB.InternalVpcInfo,
						secgroup string,
						pod *commonPB.InternalComputeInfo,
						ctx context.Context,
						vms *[]string) {
						gen.Go(func() error {
							return generateVMs(i, j, k, subnet, vpc, secgroup, pod, ctx, vms)
						})
					}(i, j, k, subnet, vpc, in.Config.VmDeploy.Secgroups[0], pod, ctx, &vms)
				}
			}
		}
		if err := gen.Wait(); err != nil {
			log.Println("Failed to generate VMs for pod ", pod.Name)
			return &pb.ReturnComputeMessage{
				ReturnMessage: "Failed to generate VMs!",
				ReturnCode:    commonPB.ReturnCode_FAILED,
				Vms:           returnVMs,
			}, err
		}
		if err := RedisClient.SAdd(
			ctx,
			constants.COMPUTE_REDIS_VM_SET,
			vms,
		).Err(); err != nil {
			return &pb.ReturnComputeMessage{
				ReturnMessage: "Failed to add vm list to redis!",
				ReturnCode:    commonPB.ReturnCode_FAILED,
				Vms:           returnVMs,
			}, err
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
		num_vms := strconv.Itoa(len(vms))
		count += len(vms)
		log.Println("Executing VM Create Workflow with VMs " + num_vms + " on pod at " + pod.ContainerIp)
		_, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, create.Create, vms, pod.ContainerIp)
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
	log.Println("Successfully started deployment for " + strconv.Itoa(count) + " VMs")
	return &pb.ReturnComputeMessage{
		ReturnMessage: "Successfully started all create workflows!",
		ReturnCode:    commonPB.ReturnCode_OK,
		Vms:           returnVMs,
	}, nil
}

func generateVMs(i, j, k int,
	subnet *commonPB.InternalSubnetInfo,
	vpc *commonPB.InternalVpcInfo,
	secgroup string,
	pod *commonPB.InternalComputeInfo,
	ctx context.Context,
	vms *[]string) error {
	vmID := pod.Id + strconv.Itoa(i) + strconv.Itoa(j) + strconv.Itoa(k)
	suffix := strconv.Itoa(i) + strconv.Itoa(j) + strconv.Itoa(k)
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
		"sg", secgroup,
		"hostIP", pod.ContainerIp,
		"hostmac", pod.Mac,
		"hostname", pod.Name,
		"status", "1",
	).Err(); err != nil {
		log.Println("Failed to hset vm ", vmID)
		return err
	}
	(*vms)[i+j+k] = vmID
	return nil
}
