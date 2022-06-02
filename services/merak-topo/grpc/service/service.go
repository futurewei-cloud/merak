package service

import (
	"context"
	"flag"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"go.temporal.io/sdk/client"
)

var (
	Port          = flag.Int("port", constants.TOPOLOGY_GRPC_SERVER_PORT, "The server port")
	returnMessage = pb.ReturnMessage{
		ReturnCode:    pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

type Server struct {
}

var TemporalClient client.Client

func (s *Server) TopologyHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnMessage, error) {
	log.Printf("Received from TopologyHandler %s", in)

	//Parse input

	switch op := in.OperationType; op {
	case INFO:
		// info return msg
	case CREATE:
		// create topology
	case DELETE:
		// delete topology
	case UPDATE:
		// update topology
	default:
		log.Println("Unknown Operation")
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "TopologyHandler: Unknown Operation"
		return &returnMessage, nil
	}

	return &in, nil
}

func (s *Server) TestHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.InternalTopologyInfo, error) {
	log.Printf("Received on TestHandler")
	returnMessage.ReturnCode = pb.ReturnCode_FAILED
	returnMessage.ReturnMessage = "Unimplemented"
	return &returnMessage, nil
}
