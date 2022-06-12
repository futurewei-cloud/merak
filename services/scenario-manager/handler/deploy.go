package handler

import (
	"errors"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/scenario-manager/database"
	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	"github.com/futurewei-cloud/merak/services/scenario-manager/utils"
)

func Deploy(s *entities.Scenario) error {
	var topology entities.TopologyConfig
	if err := database.FindEntity(s.TopologyId, utils.KEY_PREFIX_SCENARIO, &topology); err != nil {
		return errors.New("Topology not found!")
	}

	var network entities.NetworkConfig
	if err := database.FindEntity(s.NetworkConfId, utils.KEY_PREFIX_NETWORK, &network); err != nil {
		return errors.New("Network config not found!")
	}

	var service entities.ServiceConfig
	if err := database.FindEntity(s.ServiceConfId, utils.KEY_PREFIX_SERVICE, &service); err != nil {
		return errors.New("Service config not found!")
	}

	var compute entities.ComputeConfig
	if err := database.FindEntity(s.ComputeConfId, utils.KEY_PREFIX_COMPUTE, &compute); err != nil {
		return errors.New("Compute config not found!")
	}

	var topoconf pb.InternalTopologyInfo
	if err := constrcustTopologyMessage(&topology, &topoconf); err != nil {
		return errors.New("Topology protobuf message error!")
	}

	var netconf pb.InternalNetConfigInfo
	if err := constrcustNetConfMessage(&network, &netconf); err != nil {
		return errors.New("Netconfig protobuf message error!")
	}

	var computeconf pb.InternalComputeConfigInfo
	if err := constrcustComputeMessage(&compute, &computeconf); err != nil {
		return errors.New("Netconfig protobuf message error!")
	}

	//grpcclient.TopologyClient()

	return nil
}

func constrcustTopologyMessage(topo *entities.TopologyConfig, topoPb *pb.InternalTopologyInfo) error {
	topoPb.OperationType = pb.OperationType_CREATE
	topoPb.Config.FormatVersion = 1
	topoPb.Config.RevisionNumber = 1
	topoPb.Config.RequestId = utils.GenUUID()
	topoPb.Config.TopologyId = topo.Id
	topoPb.Config.MessageType = pb.MessageType_FULL
	// topoconf.TopologyType = topology.TopoType

	// repeated InternalVNodeInfo vnodes = 7;
	// repeated InternalVLinkInfo vlinks = 8;
	// InternalTopologyExtraInfo extra_info = 9;
	return nil
}

func constrcustNetConfMessage(network *entities.NetworkConfig, computePb *pb.InternalNetConfigInfo) error {

	return nil
}

func constrcustComputeMessage(compute *entities.ComputeConfig, computePb *pb.InternalComputeConfigInfo) error {

	return nil
}
