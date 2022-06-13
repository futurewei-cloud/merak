package service

import (
	"context"
	"flag"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/vm"
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
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Info Unimplemented"
		return &returnMessage, nil
	case pb.OperationType_CREATE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_CREATE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Operation Create")
		// Store Available Node IPs in DB
		for _, pod := range in.Config.Pods {
			if err := RedisClient.SAdd(ctx, "NodIPsSet", pod.Ip).Err(); err != nil {
				log.Fatalln("Unable add node IPs to DB", err)
			}
		}

		// Start VM Create Workflow
		var result string
		log.Println("Executing VM Create Workflow!")
		we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, vm.Create)
		if err != nil {
			log.Fatalln("Unable to execute workflow", err)
		}
		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

		// Synch get results of workflow
		err = we.Get(context.Background(), &result)
		if err != nil {
			log.Fatalln("Unable get workflow result", err)
		}
		log.Println("Workflow result:", result)
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Create Success!"
		return &returnMessage, nil
	case pb.OperationType_UPDATE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_UPDATE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Update Unimplemented")
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Update Unimplemented"
		return &returnMessage, nil
	case pb.OperationType_DELETE:
		workflowOptions = client.StartWorkflowOptions{
			ID:          common.VM_DELETE_WORKFLOW_ID,
			TaskQueue:   common.VM_TASK_QUEUE,
			RetryPolicy: retrypolicy,
		}
		log.Println("Delete Unimplemented")
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "ComputeHandler: Delete Unimplemented"

		return &returnMessage, nil
	default:
		log.Println("Unknown Operation")
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "ComputeHandler: Unknown Operation"
		return &returnMessage, nil
	}
}
