package service

import (
	"context"
	"flag"
	"fmt"
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
	pb.MerakTopologyServiceServer
}

func (s *Server) TopologyHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnMessage, error) {
	log.Printf("Received on TopologyHandler %s", in)

	//Parse input
	format_version := in.Config.FormatVersion
	revision_number := in.Config.RevisionNumber
	request_id := in.Config.RequestId
	topology_id := in.Config.TopologyId
	extra_info := in.Config.ExtraInfo
	vnodes := in.Config.GetVnodes()
	vlinks := in.Config.GetVlinks()
	message_type := in.Config.MessageType.Type()
	topology_type := in.Config.TopologyType.Type()

	// Operation&Return
	switch op := in.OperationType; op {
	case pb.OperationType_INFO:
		// info return msg
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Show Topology Information."
		fmt.Printf("Format_Version %v \n", format_version)
		fmt.Printf("Revision_Number %v\n", revision_number)
		fmt.Printf("Request_ID %v \n", request_id)
		fmt.Printf("Topology_ID %v \n", topology_id)
		fmt.Printf("Extra_Info %v \n", extra_info)
		fmt.Printf("Message_Type %v \n", message_type)
		fmt.Printf("Topology_Type %v \n", topology_type)
		fmt.Printf("VNode %v \n", vnodes)
		fmt.Printf("VLink %v \n", vlinks)

	case pb.OperationType_CREATE:

		// ---parse the network topology type
		// ---parse the obj types in the topology
		// ---parse the amount of pods for each objects
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
