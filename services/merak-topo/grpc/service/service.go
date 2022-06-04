package service

import (
	"context"
	"flag"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
)

var (
	Port          = flag.Int("port", constants.TOPOLOGY_GRPC_SERVER_PORT, "The server port")
	returnMessage = pb.ReturnMessage{
		ReturnCode:    pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

type Server struct {
	pb.UnimplementedMerakTopologyServiceServer
}

func (s *Server) TopologyHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnMessage, error) {
	log.Printf("Received on TopologyHandler %s", in)

	//Parse input

	switch op := in.OperationType; op {
	case pb.OperationType_INFO:
		// info return msg
	case pb.OperationType_CREATE:
		// create topology
	case pb.OperationType_DELETE:
		// delete topology
	case pb.OperationType_UPDATE:
		// update topology
	default:
		log.Println("Unknown Operation")
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "TopologyHandler: Unknown Operation"
		return &pb.ReturnMessage{}, nil
	}

	return &pb.ReturnMessage{}, nil
}

func (s *Server) TestHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnMessage, error) {
	log.Printf("Received on TestHandler %s", in)
	returnMessage.ReturnCode = pb.ReturnCode_FAILED
	returnMessage.ReturnMessage = "Unimplemented"
	return &pb.ReturnMessage{}, nil
}
