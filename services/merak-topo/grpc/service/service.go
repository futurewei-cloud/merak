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

// func (s *Server) TestHandler(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnMessage, error) {
// 	log.Printf("Received on TestHandler %s", in)
// 	returnMessage.ReturnCode = pb.ReturnCode_OK
// 	returnMessage.ReturnMessage = "Unimplemented"
// 	return &returnMessage, nil
// }

func (s *Server) TopologyHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnMessage, error) {
	log.Printf("Received on TopologyHandler %s", in)

	// Operation&Return
	switch op := in.OperationType; op {
	case pb.OperationType_INFO:
		//Parse input
		// format_version := in.Config.FormatVersion
		// revision_number := in.Config.RevisionNumber
		// request_id := in.Config.RequestId
		// topology_id := in.Config.TopologyId
		// extra_info := in.Config.ExtraInfo
		// vnodes := in.Config.GetVnodes()
		// vlinks := in.Config.GetVlinks()
		// message_type := in.Config.MessageType.String()
		// topology_type := in.Config.TopologyType.String()
		// info return msg
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Show Topology Information."
		fmt.Println("######## OperationType_INFO  Show Topology Information ########")
		fmt.Printf("Format_Version: %v \n", in.Config.FormatVersion)
		fmt.Printf("Revision_Number: %v\n", in.Config.RevisionNumber)
		fmt.Printf("Request_ID: %v \n", in.Config.RequestId)
		fmt.Printf("Topology_ID: %v \n", in.Config.TopologyId)
		fmt.Printf("Extra_Info: %v \n", in.Config.ExtraInfo)
		fmt.Printf("Message_Type: %v \n", in.Config.MessageType.String())
		fmt.Printf("Topology_Type: %v \n", in.Config.TopologyType.String())
		fmt.Printf("VNode: %v \n", in.Config.GetVnodes())
		fmt.Printf("VLink: %v \n", in.Config.GetVlinks())
		fmt.Printf("################################################################\n")

	case pb.OperationType_CREATE:

		// ---parse the network topology type
		// ---parse the obj types in the topology
		// ---parse the amount of pods for each objects

		var vhost_num = 20
		var vswitch_num = 2
		var vgateway_num = 4
		var vrouter_num = 0
		var vhost_per_vswitch = vhost_num / vswitch_num

		switch s := in.Config.TopologyType; s {
		case pb.TopologyType_TREE:

		case pb.TopologyType_SINGLE:
			//
		case pb.TopologyType_LINEAR:
			//
		case pb.TopologyType_MESH:
			//
		case pb.TopologyType_CUSTOM:
			//
		case pb.TopologyType_REVERSED:
			//
		default:
			log.Println("Unknown Topology Type ")
			returnMessage.ReturnCode = pb.ReturnCode_FAILED
			returnMessage.ReturnMessage = "TopologyHandler: Unknown Typology Type"
			return &pb.ReturnMessage{}, nil
		}

		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Topology Deployed."
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
