/*
MIT License
Copyright(c) 2022 Futurewei Cloud
    Permission is hereby granted,
    free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
    including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
    to whom the Software is furnished to do so, subject to the following conditions:
    The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
    WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package handler

import (
	"errors"
	"fmt"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/scenario-manager/database"
	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	"github.com/futurewei-cloud/merak/services/scenario-manager/grpcclient"
	"github.com/futurewei-cloud/merak/services/scenario-manager/logger"
	"github.com/futurewei-cloud/merak/services/scenario-manager/utils"
)

func TopologyHandler(s *entities.Scenario, action entities.EventName) (*pb.ReturnTopologyMessage, error) {
	var topology entities.TopologyConfig
	if err := database.FindEntity(s.TopologyId, utils.KEY_PREFIX_TOPOLOGY, &topology); err != nil {
		return nil, fmt.Errorf("topology %s not found", s.TopologyId)
	}

	if action == entities.EVENT_DEPLOY && topology.Status != entities.STATUS_NONE {
		return nil, fmt.Errorf("topology '%s' is '%s' now", topology.Id, topology.Status)
	}

	if action == entities.EVENT_DELETE {
		if topology.Status != entities.STATUS_FAILED && topology.Status != entities.STATUS_READY {
			return nil, fmt.Errorf("topology '%s' is '%s' now", topology.Id, topology.Status)
		}
	}

	if action != entities.EVENT_CHECK {
		var network entities.NetworkConfig
		if err := database.FindEntity(s.NetworkConfId, utils.KEY_PREFIX_NETWORK, &network); err != nil {
			return nil, fmt.Errorf("network config '%s' not found", s.NetworkConfId)
		}

		if network.Status != entities.STATUS_NONE {
			return nil, fmt.Errorf("nework config '%s' is '%s' now", s.NetworkConfId, network.Status)
		}

		var compute entities.ComputeConfig
		if err := database.FindEntity(s.ComputeConfId, utils.KEY_PREFIX_COMPUTE, &compute); err != nil {
			return nil, fmt.Errorf("compute config '%s' not found", s.ComputeConfId)
		}

		if compute.Status != entities.STATUS_NONE {
			return nil, fmt.Errorf("nework config '%s' is '%s' now", s.ComputeConfId, compute.Status)
		}
	}

	var topoconf pb.InternalTopologyInfo
	if err := constructTopologyMessage(&topology, &topoconf, action); err != nil {
		return nil, errors.New("topology protobuf message error")
	}
	logger.Log.Infof("constructTopologyMessage: %s", &topoconf)

	if action != entities.EVENT_CHECK {
		topology.Status = actionToStatus(action)
		database.Set(utils.KEY_PREFIX_TOPOLOGY+topology.Id, &topology)
	}

	responseTopo, err := grpcclient.TopologyClient(&topoconf)

	if err != nil || responseTopo.ReturnCode == pb.ReturnCode_FAILED {
		if action != entities.EVENT_CHECK {
			topology.Status = entities.STATUS_FAILED
			database.Set(utils.KEY_PREFIX_TOPOLOGY+topology.Id, &topology)
		}
		if responseTopo != nil {
			return nil, fmt.Errorf("deploy topology failed, return = '%s'", responseTopo.ReturnMessage)
		}
		return nil, fmt.Errorf("deploy topology failed, Error = '%s'", err.Error())
	}
	logger.Log.Infof("responseTopoMessage: %s", responseTopo)

	if action == entities.EVENT_DEPLOY {
		topology.Status = entities.STATUS_READY
	} else if action == entities.EVENT_DELETE {
		topology.Status = entities.STATUS_NONE
	}

	if action != entities.EVENT_CHECK {
		database.Set(utils.KEY_PREFIX_TOPOLOGY+topology.Id, &topology)
	}

	return responseTopo, nil
}

func actionToStatus(action entities.EventName) entities.ServiceStatus {
	switch action {
	case entities.EVENT_DEPLOY:
		return entities.STATUS_DEPLOYING
	case entities.EVENT_DELETE:
		return entities.STATUS_DELETING
	case entities.EVENT_UPDATE:
		return entities.STATUS_UPDATING
	}
	return entities.STATUS_FAILED
}

func NetworkHandler(s *entities.Scenario, action entities.EventName) (*pb.ReturnNetworkMessage, error) {
	var network entities.NetworkConfig
	if err := database.FindEntity(s.NetworkConfId, utils.KEY_PREFIX_NETWORK, &network); err != nil {
		return nil, fmt.Errorf("network config %s not found", s.NetworkConfId)
	}

	if action == entities.EVENT_DEPLOY && network.Status != entities.STATUS_NONE {
		return nil, fmt.Errorf("network '%s' is '%s' now", network.Id, network.Status)
	}

	if action == entities.EVENT_DELETE {
		if network.Status != entities.STATUS_FAILED && network.Status != entities.STATUS_READY {
			return nil, fmt.Errorf("network '%s' is '%s' now", network.Id, network.Status)
		}
	}

	var netconf pb.InternalNetConfigInfo
	if action == entities.EVENT_CHECK {
		if err := constructNetConfMessage(&network, nil, nil, &netconf, action); err != nil {
			return nil, errors.New("netconfig protobuf message error")
		}
	} else {
		var compute entities.ComputeConfig
		if err := database.FindEntity(s.ComputeConfId, utils.KEY_PREFIX_COMPUTE, &compute); err != nil {
			return nil, fmt.Errorf("compute config '%s' not found", s.ComputeConfId)
		}

		if compute.Status != entities.STATUS_NONE {
			return nil, fmt.Errorf("nework config '%s' is '%s' now", s.ComputeConfId, compute.Status)
		}

		var service entities.ServiceConfig
		if err := database.FindEntity(s.ServiceConfId, utils.KEY_PREFIX_SERVICE, &service); err != nil {
			return nil, fmt.Errorf("service %s config not found", s.ServiceConfId)
		}

		var returnTopo *pb.ReturnTopologyMessage
		returnTopo, err := TopologyHandler(s, entities.EVENT_CHECK)
		if err != nil {
			return nil, fmt.Errorf("topology %s didn't return message", s.TopologyId)
		}

		if returnTopo.ReturnCode != pb.ReturnCode_OK {
			return nil, fmt.Errorf("topology %s is not ready", s.TopologyId)
		}

		if err := constructNetConfMessage(&network, &service, returnTopo, &netconf, action); err != nil {
			return nil, errors.New("netconfig protobuf message error")
		}
	}

	logger.Log.Infof("constructNetConfMessage: %s", &netconf)

	if action != entities.EVENT_CHECK {
		network.Status = actionToStatus(action)
		database.Set(utils.KEY_PREFIX_NETWORK+network.Id, &network)
	}

	responseNetwork, err := grpcclient.NetworkClient(&netconf)

	if err != nil || responseNetwork.ReturnCode == pb.ReturnCode_FAILED {
		if action != entities.EVENT_CHECK {
			network.Status = entities.STATUS_FAILED
			database.Set(utils.KEY_PREFIX_NETWORK+network.Id, &network)
		}
		if responseNetwork != nil {
			return nil, fmt.Errorf("deploy network failed, return = '%s'", responseNetwork.ReturnMessage)
		}
		return nil, fmt.Errorf("deploy network failed, Error = '%s'", err.Error())
	}

	logger.Log.Infof("responseNetworkMessage: %s", responseNetwork)

	if action == entities.EVENT_DEPLOY {
		network.Status = entities.STATUS_READY
	} else if action == entities.EVENT_DELETE {
		network.Status = entities.STATUS_NONE
	}

	if action != entities.EVENT_CHECK {
		database.Set(utils.KEY_PREFIX_NETWORK+network.Id, &network)
	}

	return responseNetwork, nil
}

func ComputeHanlder(s *entities.Scenario, action entities.EventName) (*pb.ReturnComputeMessage, error) {
	var compute entities.ComputeConfig
	if err := database.FindEntity(s.ComputeConfId, utils.KEY_PREFIX_COMPUTE, &compute); err != nil {
		return nil, fmt.Errorf("compute config %s not found", s.ComputeConfId)
	}

	if compute.Status != entities.STATUS_NONE {
		return nil, fmt.Errorf("compute config '%s' is '%s' now", compute.Id, compute.Status)
	}

	if action == entities.EVENT_DEPLOY && compute.Status != entities.STATUS_NONE {
		return nil, fmt.Errorf("compute '%s' is '%s' now", compute.Id, compute.Status)
	}

	if action == entities.EVENT_DELETE {
		if compute.Status != entities.STATUS_FAILED && compute.Status != entities.STATUS_READY {
			return nil, fmt.Errorf("compute '%s' is '%s' now)", compute.Id, compute.Status)
		}
	}

	var computeconf pb.InternalComputeConfigInfo
	if action == entities.EVENT_CHECK {
		if err := constructComputeMessage(&compute, nil, nil, nil, &computeconf, action); err != nil {
			return nil, errors.New("netconfig protobuf message error")
		}
	} else {
		var service entities.ServiceConfig
		if err := database.FindEntity(s.ServiceConfId, utils.KEY_PREFIX_SERVICE, &service); err != nil {
			return nil, errors.New("service config not found")
		}

		var returnTopo *pb.ReturnTopologyMessage
		returnTopo, topo_err := TopologyHandler(s, entities.EVENT_CHECK)
		if topo_err != nil {
			return nil, fmt.Errorf("topology %s didn't return message", s.TopologyId)
		}

		if returnTopo.ReturnCode != pb.ReturnCode_OK {
			return nil, fmt.Errorf("topology %s is not ready", s.TopologyId)
		}

		var returnNetwork *pb.ReturnNetworkMessage
		returnNetwork, net_err := NetworkHandler(s, entities.EVENT_CHECK)
		if net_err != nil {
			return nil, fmt.Errorf("network %s didn't return message", s.NetworkConfId)
		}

		if returnNetwork.ReturnCode != pb.ReturnCode_OK {
			return nil, fmt.Errorf("network %s is not ready", s.NetworkConfId)
		}

		if err := constructComputeMessage(&compute, &service, returnTopo, returnNetwork, &computeconf, action); err != nil {
			return nil, errors.New("netconfig protobuf message error")
		}

		for _, n := range returnNetwork.GetVpcs() {
			for _, s := range n.GetSubnets() {
				if compute.NumberOfComputeNodes != 0 && len(n.GetSubnets()) != 0 {
					s.NumberVms = uint32(compute.NumberOfVmPerVpc) / uint32(compute.NumberOfComputeNodes) / uint32(len(n.GetSubnets()))
				}
			}
		}
		logger.Log.Infof("after returnNetworkMessage: %s", returnNetwork)
	}

	logger.Log.Infof("constructComputeMessage: %s", &computeconf)

	if action != entities.EVENT_CHECK {
		compute.Status = actionToStatus(action)
		database.Set(utils.KEY_PREFIX_COMPUTE+compute.Id, &compute)
	}

	responseCompute, err := grpcclient.ComputeClient(&computeconf)

	if err != nil || (responseCompute != nil && responseCompute.ReturnCode == pb.ReturnCode_FAILED) {
		if action != entities.EVENT_CHECK {
			compute.Status = entities.STATUS_FAILED
			database.Set(utils.KEY_PREFIX_COMPUTE+compute.Id, &compute)
		}
		if responseCompute != nil {
			return nil, fmt.Errorf("deploy compute failed, return = '%s'", responseCompute.ReturnMessage)
		}
		return nil, fmt.Errorf("deploy compute failed, Error = '%s'", err.Error())
	}

	logger.Log.Infof("responseComputeMessage: %s", responseCompute)

	if action == entities.EVENT_DEPLOY {
		compute.Status = entities.STATUS_READY
	} else if action == entities.EVENT_DELETE {
		compute.Status = entities.STATUS_NONE
	}

	if action != entities.EVENT_CHECK {
		database.Set(utils.KEY_PREFIX_NETWORK+compute.Id, &compute)
	}

	return responseCompute, nil
}
