package handler

import (
	"context"
	"errors"
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
		return &pb.ReturnMessage{
			ReturnMessage: "Info Unimplemented",
			ReturnCode:    pb.ReturnCode_FAILED,
		}, errors.New("info unimplemented")

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
				return &pb.ReturnMessage{
					ReturnMessage: "Unable add pod to DB Hash Map",
					ReturnCode:    pb.ReturnCode_FAILED,
				}, err
			}
			log.Println("Added pod " + pod.Name + " at address " + pod.Ip)
			if err := RedisClient.SAdd(
				ctx,
				constants.COMPUTE_REDIS_NODE_IP_SET,
				pod.Id,
			).Err(); err != nil {
				return &pb.ReturnMessage{
					ReturnMessage: "Unable add pod to DB Hash Set",
					ReturnCode:    pb.ReturnCode_FAILED,
				}, err
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
							"subnetID", subnet.SubnetId,
							"cidr", subnet.SubnetCidr,
							"gw", subnet.SubnetGw,
							"sg", in.Config.VmDeploy.Secgroups[0],
							"hostIP", pod.Ip,
							"hostmac", pod.Mac,
							"hostname", pod.Name,
						).Err(); err != nil {
							return &pb.ReturnMessage{
								ReturnMessage: "Unable add VM to DB Hash Map",
								ReturnCode:    pb.ReturnCode_FAILED,
							}, err
						}
						log.Println("Added VM " + pod.Id + strconv.Itoa(j) + " for vpc " + vpc.VpcId + " for subnet " + subnet.SubnetId + " vm number " + strconv.Itoa(j) + " of " + strconv.Itoa(int(subnet.NumberVms)))
						if err := RedisClient.LPush(ctx, pod.Id, pod.Id+strconv.Itoa(j)).Err(); err != nil {
							return &pb.ReturnMessage{
								ReturnMessage: "Unable add VM to pod list",
								ReturnCode:    pb.ReturnCode_FAILED,
							}, err
						}
						log.Println("Added pod -> vm mapping " + pod.Id + strconv.Itoa(j))
					}
				}
			}

		}
		// Start VM Create Workflow
		var result pb.ReturnMessage
		log.Println("Executing VM Create Workflow!")
		we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, create.Create)
		if err != nil {
			return &pb.ReturnMessage{
				ReturnMessage: "Unable to execute workflow",
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		log.Println("Started workflow WorkflowID "+we.GetID()+" RunID ", we.GetRunID())

		// Sync get results of workflow
		err = we.Get(context.Background(), &result)
		if err != nil {
			return &pb.ReturnMessage{
				ReturnMessage: result.GetReturnMessage(),
				ReturnCode:    pb.ReturnCode_FAILED,
			}, err
		}
		log.Println("Workflow result:", result.ReturnMessage)
		return &pb.ReturnMessage{
			ReturnMessage: result.GetReturnMessage(),
			ReturnCode:    result.GetReturnCode(),
		}, err

	case pb.OperationType_UPDATE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_UPDATE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Update Unimplemented")
		return &pb.ReturnMessage{
			ReturnMessage: "Update Unimplemented",
			ReturnCode:    pb.ReturnCode_FAILED,
		}, errors.New("update unimplemented")

	case pb.OperationType_DELETE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_DELETE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Delete Unimplemented")
		return &pb.ReturnMessage{
			ReturnMessage: "Delete Unimplemented",
			ReturnCode:    pb.ReturnCode_FAILED,
		}, errors.New("delete unimplemented")

	default:
		log.Println("Unknown Operation")
		return &pb.ReturnMessage{
			ReturnMessage: "Unknown Operation",
			ReturnCode:    pb.ReturnCode_FAILED,
		}, errors.New("unknown operation")
	}
}
