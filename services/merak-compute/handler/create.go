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
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/create"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

func caseCreate(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnComputeMessage, error) {

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_WF_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_WF_BACKOFF,
		MaximumInterval:    common.TEMPORAL_WF_MAX_INTERVAL,
		MaximumAttempts:    common.TEMPORAL_WF_MAX_ATTEMPT,
	}

	log.Println("Operation Create")
	return_vms := []*pb.InternalVMInfo{}
	// Add pods to DB
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
				ReturnCode:    common_pb.ReturnCode_FAILED,
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
				ReturnCode:    common_pb.ReturnCode_FAILED,
			}, err
		}

		// Generate VMs for each VPC and Subnet
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
							ReturnCode:    common_pb.ReturnCode_FAILED,
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
							ReturnCode:    common_pb.ReturnCode_FAILED,
						}, err
					}
					return_vm := pb.InternalVMInfo{
						Id:              vmID,
						Name:            "v" + suffix,
						VpcId:           vpc.VpcId,
						Ip:              "",
						SecurityGroupId: in.Config.VmDeploy.Secgroups[0],
						SubnetId:        subnet.SubnetId,
						DefaultGateway:  subnet.SubnetGw,
						Status:          common_pb.Status(1),
					}
					return_vms = append(return_vms, &return_vm)
					// Store VM to Pod list
					log.Println("Added VM " + vmID + " for vpc " + vpc.VpcId + " for subnet " + subnet.SubnetId + " vm number " + strconv.Itoa(k+1) + " of " + strconv.Itoa(int(subnet.NumberVms)))
					if err := RedisClient.LPush(ctx, "l"+pod.Id, vmID).Err(); err != nil {
						log.Println("Failed to add pod -> vm mapping " + vmID)
						return &pb.ReturnComputeMessage{
							ReturnMessage: "Unable add VM to pod list",
							ReturnCode:    common_pb.ReturnCode_FAILED,
							Vms:           return_vms,
						}, err
					}
					log.Println("Added pod -> vm mapping " + vmID)
				}
			}
		}
		// Get VM to pod list
		vms := RedisClient.LRange(ctx, "l"+pod.Id, 0, -1)
		if vms.Err() != nil {
			log.Println("Unable get node vmIDsList from redis", vms.Err())
			return &pb.ReturnComputeMessage{
				ReturnCode:    common_pb.ReturnCode_FAILED,
				ReturnMessage: "Unable get node vmIDsList from redis",
				Vms:           return_vms,
			}, vms.Err()
		}

		// Execute VM creation on a per pod basis
		// Send a list of VMs to the Workflow
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_CREATE_WORKFLOW_ID + strconv.Itoa(n),
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Executing VM Create Workflow with VMs ", vms.Val())
		we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, create.Create, vms.Val())
		if err != nil {
			return &pb.ReturnComputeMessage{
				ReturnMessage: "Unable to execute create workflow",
				ReturnCode:    common_pb.ReturnCode_FAILED,
				Vms:           return_vms,
			}, err
		}
		log.Println("Started Create workflow WorkflowID "+we.GetID()+" RunID ", we.GetRunID())
	}

	return &pb.ReturnComputeMessage{
		ReturnMessage: "Successfully started all create workflows!",
		ReturnCode:    common_pb.ReturnCode_OK,
		Vms:           return_vms,
	}, nil
}
