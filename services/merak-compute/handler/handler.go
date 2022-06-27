package handler

import (
	"context"
	"flag"
	"log"
	"strconv"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	create "github.com/futurewei-cloud/merak/services/merak-compute/workflows"
	"github.com/go-redis/redis/v9"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

var (
	Port            = flag.Int("port", constants.COMPUTE_GRPC_SERVER_PORT, "The server port")
	workflowOptions client.StartWorkflowOptions
	returnMessage   = pb.ReturnMessage{
		ReturnCode:    pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

var TemporalClient client.Client
var RedisClient redis.Client

type Server struct {
	pb.UnimplementedMerakComputeServiceServer
}

func (s *Server) ComputeHandler(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnMessage, error) {
	log.Println("Received on ComputeHandler", in)
	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    common.TEMPORAL_RETRY_INTERVAL,
		BackoffCoefficient: common.TEMPORAL_BACKOFF,
		MaximumInterval:    common.TEMPORAL_MAX_INTERVAL,
	}

	// Parse input
	switch op := in.OperationType; op {
	case pb.OperationType_INFO:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_INFO_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Info Unimplemented")
		SetReturnMessage("Info Unimplemented", pb.ReturnCode_FAILED)
		return &returnMessage, nil

	case pb.OperationType_CREATE:
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
				"ip", pod.Ip,
				"mac", pod.Mac,
				"veth", pod.Veth,
			).Err(); err != nil {
				SetReturnMessage("Unable add pod to DB Hash Map", pb.ReturnCode_FAILED)
				return &returnMessage, err
			}
			log.Println("Added", "pod", pod.Name)
			if err := RedisClient.SAdd(
				ctx,
				constants.COMPUTE_REDIS_NODE_IP_SET,
				pod.Id,
			).Err(); err != nil {
				SetReturnMessage("Unable add pod to DB Hash Set", pb.ReturnCode_FAILED)
				return &returnMessage, err
			}

			// Currently 1 VM = 1 Port.
			// Generate VMs
			for _, vpc := range in.Config.VmDeploy.Vpcs {
				for _, subnet := range vpc.Subnets {
					for j := 0; j < int(subnet.NumberVms); j++ {
						if err := RedisClient.HSet(
							ctx,
							pod.Id+strconv.Itoa(j),
							"id", pod.Id+strconv.Itoa(j),
							"name", pod.Id+strconv.Itoa(j),
							"vpc", vpc.VpcId,
							"tenantID", vpc.TenantId,
							"projectID", vpc.ProjectId,
							"subnet", subnet.SubnetId,
							"gw", subnet.SubnetGw,
							"sg", in.Config.VmDeploy.Secgroups[0],
						).Err(); err != nil {
							SetReturnMessage("Unable add VM to DB Hash Map", pb.ReturnCode_FAILED)
							return &returnMessage, err
						}
						log.Println("Added", "VM", pod.Id+strconv.Itoa(j), "vpc", vpc.VpcId, "subnet", subnet.SubnetId, "vm number", j, "of", int(subnet.NumberVms))
						if err := RedisClient.LPush(ctx, pod.Id, pod.Id+strconv.Itoa(j)).Err(); err != nil {
							SetReturnMessage("Unable add VM to pod list", pb.ReturnCode_FAILED)
							return &returnMessage, err
						}
						log.Println("Added pod -> vm mapping ", pod.Id+strconv.Itoa(j))
					}
				}
			}

		}
		// Start VM Create Workflow
		var result string
		log.Println("Executing VM Create Workflow!")
		we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, create.Create)
		if err != nil {
			SetReturnMessage("Unable to execute workflow", pb.ReturnCode_FAILED)
			return &returnMessage, err
		}
		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

		// Sync get results of workflow
		err = we.Get(context.Background(), &result)
		if err != nil {
			SetReturnMessage("Unable get workflow result", pb.ReturnCode_FAILED)
			return &returnMessage, err
		}
		log.Println("Workflow result:", result)
		SetReturnMessage("Create Success!", pb.ReturnCode_OK)
		return &returnMessage, nil

	case pb.OperationType_UPDATE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_UPDATE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Update Unimplemented")
		SetReturnMessage("Update Unimplemented", pb.ReturnCode_FAILED)
		return &returnMessage, nil

	case pb.OperationType_DELETE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_DELETE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Delete Unimplemented")
		SetReturnMessage("Delete Unimplemented", pb.ReturnCode_FAILED)
		return &returnMessage, nil

	default:
		log.Println("Unknown Operation")
		SetReturnMessage("ComputeHandler: Unknown Operation", pb.ReturnCode_FAILED)
		return &returnMessage, nil
	}
}

func SetReturnMessage(returnString string, returnCode pb.ReturnCode) {
	returnMessage.ReturnCode = returnCode
	returnMessage.ReturnMessage = returnString
}
