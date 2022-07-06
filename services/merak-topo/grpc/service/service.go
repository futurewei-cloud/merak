package service

import (
	"context"
	"flag"
	"fmt"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	"github.com/futurewei-cloud/merak/services/merak-topo/handler"
	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
)

var (
	Port = flag.Int("port", constants.TOPLOGY_GRPC_SERVER_PORT, "The server port")

	returnMessage = pb.ReturnTopologyMessage{
		ReturnCode:    pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

type Server struct {
	pb.MerakTopologyServiceServer
}

func (s *Server) TopologyHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	log.Printf("Received on TopologyHandler %s", in)

	k8client, err := utils.K8sClient() // config.yaml- sm
	if err != nil {
		return nil, fmt.Errorf("create k8s client error %s", err.Error())
	}

	err1 := database.ConnectDatabase()
	if err1 != nil {
		fmt.Printf("connect to DB error %s", err1)
	}

	// Operation&Return
	switch op := in.OperationType; op {
	case pb.OperationType_INFO:

		// query  based on topology_id

		if in.Config.GetTopologyId() != "" {
			err_info := handler.Info(k8client, in.Config.GetTopologyId(), &returnMessage)
			if err_info != nil {
				returnMessage.ReturnCode = pb.ReturnCode_FAILED
				returnMessage.ReturnMessage = "INFO fails to query on topology id."

			} else {
				returnMessage.ReturnCode = pb.ReturnCode_OK
				returnMessage.ReturnMessage = "INFO successes to query on topology id."
			}
		}

	case pb.OperationType_CREATE:

		/// save the protobuf hostinfo --- with each created topology

		aca_num := in.Config.GetNumberOfVhosts()
		rack_num := in.Config.GetNumberOfRacks()
		aca_per_rack := in.Config.GetVhostPerRack()
		data_plane_cidr := in.Config.GetDataPlaneCidr()
		topo_id := in.Config.GetTopologyId()

		if data_plane_cidr == "" || aca_num == 0 || aca_per_rack == 0 || rack_num == 0 {

			returnMessage.ReturnCode = pb.ReturnCode_FAILED
			returnMessage.ReturnMessage = "Must provide a valid data plane cider, aca number, aca per rack number and rack number"
			// returnMessage.ComputeNodes = _nodes

			return &returnMessage, nil

		}

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
			err_create := handler.Create(k8client, topo_id, uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), data_plane_cidr, &returnMessage)

			//return topology message-- compute info

			if err_create != nil {
				returnMessage.ReturnCode = pb.ReturnCode_FAILED
				returnMessage.ReturnMessage = "Fail to Create Topology."

			} else {
				returnMessage.ReturnCode = pb.ReturnCode_OK
				returnMessage.ReturnMessage = "Success to create topology"

			}
			return &returnMessage, err_create

		}

	case pb.OperationType_DELETE:
		// delete topology
		err := handler.Delete(k8client, in.Config.TopologyId)

		//return topology message-- compute info

		if err != nil {
			returnMessage.ReturnCode = pb.ReturnCode_FAILED
			returnMessage.ReturnMessage = "Fail to Delete Topology."
		} else {
			returnMessage.ReturnCode = pb.ReturnCode_OK
			returnMessage.ReturnMessage = "Success to Delete Topology"
		}

	case pb.OperationType_UPDATE:
		// update topology
	default:
		log.Println("Unknown Operation")
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "TopologyHandler: Unknown Operation"

	}

	return &returnMessage, nil
}

func (s *Server) TestHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.ReturnTopologyMessage, error) {
	log.Printf("Received on TopologyHandler %s", in)
	return &returnMessage, nil
}
