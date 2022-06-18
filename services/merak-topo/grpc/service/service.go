package service

import (
	"context"
	"flag"
	"fmt"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-topo/handler"
)

var (
	Port          = flag.Int("port", constants.TOPLOGY_GRPC_SERVER_PORT, "The server port")
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

func (s *Server) TopologyHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	log.Printf("Received on TopologyHandler %s", in)

	// Operation&Return
	switch op := in.OperationType; op {
	case pb.OperationType_INFO:
		//Parse input
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Show Topology Information."
		fmt.Println("######## OperationType_INFO  Show Topology Information ########")
		fmt.Printf("Format_Version: %v \n", in.Config.FormatVersion)
		fmt.Printf("Revision_Number: %v\n", in.Config.RevisionNumber)
		fmt.Printf("Request_ID: %v \n", in.Config.RequestId)
		fmt.Printf("Message_Type: %v \n", in.Config.MessageType.String())
		fmt.Printf("Topology_ID: %v \n", in.Config.TopologyId)
		fmt.Printf("Name: %v \n", in.Config.Name)
		fmt.Printf("Topology_Type: %v \n", in.Config.TopologyType.String())
		fmt.Printf("Number of Vhosts: %v \n", in.Config.NumberOfVhosts)
		fmt.Printf("Number of Racks: %v \n", in.Config.NumberOfRacks)
		fmt.Printf("Vhosts Per Rack: %v \n", in.Config.VhostPerRack)
		fmt.Printf("Data Plane CIDR: %v \n", in.Config.DataPlaneCidr)
		fmt.Printf("Number of Gateway: %v \n", in.Config.NumberOfGateways)
		fmt.Printf("Gateway IPs: %v \n", in.Config.GetGatewayIps())
		fmt.Printf("Images: %v \n", in.Config.GetImages())
		fmt.Printf("VNode: %v \n", in.Config.GetVnodes())
		fmt.Printf("VLink: %v \n", in.Config.GetVlinks())
		fmt.Printf("Extra_Info: %v \n", in.Config.ExtraInfo)
		fmt.Printf("################################################################\n")

	case pb.OperationType_CREATE:

		aca_num := 8
		rack_num := 4
		aca_per_rack := 2
		data_plane_cidr := "10.200.0.0/16"

		// save topology data to radis

		switch s := in.Config.TopologyType; s {
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
			// pb.TopologyType_TREE
			topology_create := handler.Create(uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), data_plane_cidr)
			fmt.Printf("The created topology is: %+v \n", topology_create)

		}

		// returnMessage.ReturnCode = pb.ReturnCode_OK
		// returnMessage.ReturnMessage = "Topology Deployed."
	case pb.OperationType_DELETE:
		// delete topology

	case pb.OperationType_UPDATE:
		// update topology
	default:
		log.Println("Unknown Operation")
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "TopologyHandler: Unknown Operation"
		return &pb.ReturnTopologyMessage{}, nil
	}

	return &pb.ReturnTopologyMessage{}, nil
}

func (s *Server) TestHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	log.Printf("Received on TopologyHandler %s", in)
	return &pb.ReturnTopologyMessage{}, nil
}
