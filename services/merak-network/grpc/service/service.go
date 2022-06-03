package service

import (
	"context"
	"flag"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	virtualnetwork "github.com/futurewei-cloud/merak/services/merak-network/workflows/virtual_network"
	"go.temporal.io/sdk/client"
)

var (
	Port            = flag.Int("port", constants.NETWORK_GRPC_SERVER_PORT, "The server port")
	workflowOptions client.StartWorkflowOptions
	returnMessage   = pb.ReturnMessage{
		ReturnCode:    pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

type Server struct {
	pb.UnimplementedMerakNetworkServiceServer
}

var TemporalClient client.Client

func (s *Server) NetworkHandler(ctx context.Context, in *pb.InternalNetConfigInfo) (*pb.ReturnMessage, error) {
	log.Printf("Received on NetworkHandler %s", in)
	// Parse input

	switch op := in.OperationType; op {
	case pb.OperationType_INFO:
		workflowOptions = client.StartWorkflowOptions{
			ID:        "InfoWorkflow",
			TaskQueue: "Info",
		}
	case pb.OperationType_CREATE:
		workflowOptions = client.StartWorkflowOptions{
			ID:        "CreateWorkflow",
			TaskQueue: "Create",
		}
		for _, services := range in.Configuration.Services {
			log.Println(services)
		}
		for _, compute := range in.Configuration.Computes {
			log.Println(compute)
		}
		for _, network := range in.Configuration.Networks {
			log.Println(network)
		}
		for _, storage := range in.Configuration.Storages {
			log.Println(storage)
		}
		for _, extraInfo := range in.Configuration.ExtraInfo {
			log.Println(extraInfo)
		}
	case pb.OperationType_UPDATE:
		workflowOptions = client.StartWorkflowOptions{
			ID:        "UpdateWorkflow",
			TaskQueue: "Update",
		}
	case pb.OperationType_DELETE:
		workflowOptions = client.StartWorkflowOptions{
			ID:        "DeleteWorkflow",
			TaskQueue: "Delete",
		}
	default:
		log.Println("Unknown Operation")
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "NetworkHandler: Unknown Operation"
		return &returnMessage, nil
	}

	we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, virtualnetwork.Create, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	return &returnMessage, nil
}
