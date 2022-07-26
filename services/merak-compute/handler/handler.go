package handler

import (
	"context"
	"errors"
	"flag"
	"log"
	"strconv"

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	compute_pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	create "github.com/futurewei-cloud/merak/services/merak-compute/workflows/create"
	"github.com/go-redis/redis/v9"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

var (
	Port            = flag.Int("port", constants.COMPUTE_GRPC_SERVER_PORT, "The server port")
	workflowOptions client.StartWorkflowOptions
	returnMessage   = compute_pb.ReturnMessage{
		ReturnCode:    common_pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

var TemporalClient client.Client
var RedisClient redis.Client

type Server struct {
	compute_pb.UnimplementedMerakComputeServiceServer
}

func (s *Server) ComputeHandler(ctx context.Context, in *compute_pb.InternalComputeConfigInfo) (*compute_pb.ReturnMessage, error) {
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
		var result compute_pb.ReturnMessage
		log.Println("Executing VM Info Workflow!")
		we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, create.Create)
		if err != nil {
			return &compute_pb.ReturnMessage{
				ReturnMessage: "Unable to execute workflow",
				ReturnCode:    common_pb.ReturnCode_FAILED,
			}, err
		}
		log.Println("Started workflow WorkflowID "+we.GetID()+" RunID ", we.GetRunID())

		// Sync get results of workflow
		err = we.Get(context.Background(), &result)
		if err != nil {
			return &compute_pb.ReturnMessage{
				ReturnMessage: result.GetReturnMessage(),
				ReturnCode:    common_pb.ReturnCode_FAILED,
				ReturnVms:     result.GetReturnVms(),
			}, err
		}
		log.Println("Workflow result:", result.ReturnMessage)
		return &compute_pb.ReturnMessage{
			ReturnMessage: result.GetReturnMessage(),
			ReturnCode:    result.GetReturnCode(),
			ReturnVms:     result.GetReturnVms(),
		}, err

	case common_pb.OperationType_CREATE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_CREATE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Operation Create")
		// Store Available Node IPs in DB

		//Info needed:
		// -All host IPs(pod ips)
		// -VMs to create/host
		// -Ports to create/VM
		// Subnet and VPC for each VM

		for _, pod := range in.Config.Pods {
			if err := RedisClient.HSet(
				ctx,
				constants.COMPUTE_REDIS_NODE_MAP,
				"id", pod.Id,
				"name", pod.Name,
				"ip", pod.ContainerIp,
				"mac", pod.Mac,
				"veth", pod.Veth,
			).Err(); err != nil {
				return &compute_pb.ReturnMessage{
					ReturnMessage: "Unable add pod to DB Hash Map",
					ReturnCode:    common_pb.ReturnCode_FAILED,
				}, err
			}
			log.Println("Added pod " + pod.Name + " at address " + pod.Ip)
			if err := RedisClient.SAdd(
				ctx,
				constants.COMPUTE_REDIS_NODE_IP_SET,
				pod.Id,
			).Err(); err != nil {
				return &compute_pb.ReturnMessage{
					ReturnMessage: "Unable to add pod to DB Hash Set",
					ReturnCode:    common_pb.ReturnCode_FAILED,
				}, err
			}

			// Currently 1 VM = 1 Port.
			// Generate VMs
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
							return &compute_pb.ReturnMessage{
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
						).Err(); err != nil {
							return &compute_pb.ReturnMessage{
								ReturnMessage: "Unable add VM to DB Hash Map",
								ReturnCode:    common_pb.ReturnCode_FAILED,
							}, err
						}
						log.Println("Added VM " + vmID + " for vpc " + vpc.VpcId + " for subnet " + subnet.SubnetId + " vm number " + strconv.Itoa(i) + strconv.Itoa(j) + strconv.Itoa(k) + " of " + strconv.Itoa(int(subnet.NumberVms)))
						if err := RedisClient.LPush(ctx, pod.Id, vmID).Err(); err != nil {
							return &compute_pb.ReturnMessage{
								ReturnMessage: "Unable add VM to pod list",
								ReturnCode:    common_pb.ReturnCode_FAILED,
							}, err
						}
						log.Println("Added pod -> vm mapping " + vmID)
					}
				}
			}

		}
		// Start VM Create Workflow
		var result compute_pb.ReturnMessage
		log.Println("Executing VM Create Workflow!")
		we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, create.Create)
		if err != nil {
			return &compute_pb.ReturnMessage{
				ReturnMessage: "Unable to execute workflow",
				ReturnCode:    common_pb.ReturnCode_FAILED,
			}, err
		}
		log.Println("Started workflow WorkflowID "+we.GetID()+" RunID ", we.GetRunID())

		// Sync get results of workflow
		err = we.Get(context.Background(), &result)
		if err != nil {
			return &compute_pb.ReturnMessage{
				ReturnMessage: result.GetReturnMessage(),
				ReturnCode:    common_pb.ReturnCode_FAILED,
				ReturnVms:     result.GetReturnVms(),
			}, err
		}
		log.Println("Workflow result:", result.ReturnMessage)
		return &compute_pb.ReturnMessage{
			ReturnMessage: result.GetReturnMessage(),
			ReturnCode:    result.GetReturnCode(),
			ReturnVms:     result.GetReturnVms(),
		}, err

	case common_pb.OperationType_UPDATE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_UPDATE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Update Unimplemented")
		return &compute_pb.ReturnMessage{
			ReturnMessage: "Update Unimplemented",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, errors.New("update unimplemented")

	case common_pb.OperationType_DELETE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_DELETE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Delete Unimplemented")
		return &compute_pb.ReturnMessage{
			ReturnMessage: "Delete Unimplemented",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, errors.New("delete unimplemented")

	default:
		log.Println("Unknown Operation")
		return &compute_pb.ReturnMessage{
			ReturnMessage: "Unknown Operation",
			ReturnCode:    common_pb.ReturnCode_FAILED,
		}, errors.New("unknown operation")
	}
}
