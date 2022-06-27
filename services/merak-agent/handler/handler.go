package handler

import (
	"context"
	"flag"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
)

var (
	Port          = flag.Int("port", constants.COMPUTE_GRPC_SERVER_PORT, "The server port")
	returnMessage = pb.ReturnMessage{
		ReturnCode:    pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

type Server struct {
	pb.UnimplementedMerakAgentServiceServer
}

func (s *Server) ComputeHandler(ctx context.Context, in *pb.InternalPortConfig) (*pb.ReturnMessage, error) {
	log.Println("Received on ComputeHandler", in)

	// Parse input
	switch op := in.OperationType; op {
	case pb.OperationType_INFO:
		log.Println("Info Unimplemented")
		SetReturnMessage("Info Unimplemented", pb.ReturnCode_FAILED)
		return &returnMessage, nil

	case pb.OperationType_CREATE:
		log.Println("Operation Create")
		log.Println("pb received!", in)
		SetReturnMessage("Create Success!", pb.ReturnCode_OK)
		return &returnMessage, nil

	case pb.OperationType_UPDATE:

		log.Println("Update Unimplemented")
		SetReturnMessage("Update Unimplemented", pb.ReturnCode_FAILED)
		return &returnMessage, nil

	case pb.OperationType_DELETE:

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
