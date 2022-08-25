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
	"errors"
	"flag"
	"log"
	"strconv"

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	create "github.com/futurewei-cloud/merak/services/merak-compute/workflows/create"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/delete"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/info"
	"github.com/go-redis/redis/v9"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

var (
	Port            = flag.Int("port", constants.COMPUTE_GRPC_SERVER_PORT, "The server port")
	workflowOptions client.StartWorkflowOptions
	returnMessage   = common_pb.ReturnMessage{
		ReturnCode:    common_pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

var TemporalClient client.Client
var RedisClient redis.Client

type Server struct {
	pb.UnimplementedMerakComputeServiceServer
}

func (s *Server) ComputeHandler(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnComputeMessage, error) {
	log.Println("Received on ComputeHandler", in)
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_WF_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_WF_BACKOFF,
		MaximumInterval:    common.TEMPORAL_WF_MAX_INTERVAL,
		MaximumAttempts:    common.TEMPORAL_WF_MAX_ATTEMPT,
	}

	// Parse input
	switch op := in.OperationType; op {
	case common_pb.OperationType_INFO:
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

	case common_pb.OperationType_CREATE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_CREATE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Operation Create")

		// Add pods to DB
		for _, pod := range in.Config.Pods {
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
					for k := 0; j < int(subnet.NumberVms); j++ {
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

						// Store VM to Pod list
						log.Println("Added VM " + vmID + " for vpc " + vpc.VpcId + " for subnet " + subnet.SubnetId + " vm number " + strconv.Itoa(i) + strconv.Itoa(j) + strconv.Itoa(k) + " of " + strconv.Itoa(int(subnet.NumberVms)))
						if err := RedisClient.LPush(ctx, "l"+pod.Id, vmID).Err(); err != nil {
							return &pb.ReturnComputeMessage{
								ReturnMessage: "Unable add VM to pod list",
								ReturnCode:    common_pb.ReturnCode_FAILED,
							}, err
						}
						log.Println("Added pod -> vm mapping " + vmID)
					}
				}
			}
			// Get VM to pod list
			vms := common.RedisClient.LRange(ctx, "l"+pod.Id, 0, -1)
			if vms.Err() != nil {
				log.Println("Unable get node vmIDsList from redis", vms.Err())
				return &pb.ReturnComputeMessage{
					ReturnCode:    common_pb.ReturnCode_FAILED,
					ReturnMessage: "Unable get node vmIDsList from redis",
				}, vms.Err()
			}

			// Execute VM creation on a per pod basis
			// Send a list of VMs to the Workflow
			log.Println("Executing VM Create Workflow!")
			we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, create.Create, vms.Val())
			if err != nil {
				return &pb.ReturnComputeMessage{
					ReturnMessage: "Unable to execute create workflow",
					ReturnCode:    common_pb.ReturnCode_FAILED,
				}, err
			}
			log.Println("Started Create workflow WorkflowID "+we.GetID()+" RunID ", we.GetRunID())
		}

		return &pb.ReturnComputeMessage{
			ReturnMessage: "Successfully started all create workflows!",
			ReturnCode:    common_pb.ReturnCode_OK,
		}, nil

	case common_pb.OperationType_UPDATE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_UPDATE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Update Unimplemented")
		return &pb.ReturnComputeMessage{
			ReturnMessage: "Update Unimplemented",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, errors.New("update unimplemented")

	case common_pb.OperationType_DELETE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_DELETE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}

		//Get a list of all pods
		pod_list := common.RedisClient.SMembers(
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
		for _, pod_id := range pod_list.Val() {
			vms := common.RedisClient.LRange(ctx, "l"+pod_id, 0, -1)
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

		return &pb.ReturnComputeMessage{
			ReturnMessage: "Successfully started all delete workflows!",
			ReturnCode:    common_pb.ReturnCode_OK,
		}, nil

	default:
		log.Println("Unknown Operation")
		return &pb.ReturnComputeMessage{
			ReturnMessage: "Unknown Operation",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, errors.New("unknown operation")
	}
}
