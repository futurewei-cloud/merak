package service

import (
	"context"
	"flag"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/vm"
	"go.temporal.io/sdk/client"
)

var (
	Port = flag.Int("port", constants.COMPUTE_GRPC_SERVER_PORT, "The server port")
)

type Server struct {
	pb.UnimplementedMerakComputeServiceServer
}

var TemporalClient client.Client

func (s *Server) ComputeHandler(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.InternalComputeConfigInfo, error) {
	log.Printf("Received on ComputeHandler %s", in)
	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello_world_workflowID",
		TaskQueue: "hello-world",
	}

	we, err := TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, vm.Create, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	return in, nil
}

func (s *Server) TestHandler(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.InternalComputeConfigInfo, error) {
	log.Printf("Received on TestHandler %s", in)
	return in, nil
}
