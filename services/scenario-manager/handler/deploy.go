package handler

import (
	"errors"
	"fmt"
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/scenario-manager/database"
	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	"github.com/futurewei-cloud/merak/services/scenario-manager/grpcclient"
	"github.com/futurewei-cloud/merak/services/scenario-manager/utils"
)

func TopologyHandler(s *entities.Scenario, responseTopo *pb.ReturnTopologyMessage) error {
	var topology entities.TopologyConfig
	if err := database.FindEntity(s.TopologyId, utils.KEY_PREFIX_SCENARIO, &topology); err != nil {
		return errors.New("topology not found")
	}

	if topology.Status != entities.STATUS_NONE {
		return fmt.Errorf("topology '%s' is '%s' now)", topology.Id, topology.Status)
	}

	var network entities.NetworkConfig
	if err := database.FindEntity(s.NetworkConfId, utils.KEY_PREFIX_NETWORK, &network); err != nil {
		return fmt.Errorf("network config '%s' not found", s.NetworkConfId)
	}

	if network.Status != entities.STATUS_NONE {
		return fmt.Errorf("nework config '%s' is '%s' now", s.NetworkConfId, network.Status)
	}

	var topoconf pb.InternalTopologyInfo
	if err := constructTopologyMessage(&topology, &topoconf); err != nil {
		return errors.New("topology protobuf message error")
	}

	topology.Status = entities.STATUS_DEPLOYING
	database.Set(utils.KEY_PREFIX_TOPOLOGY+topology.Id, &topology)

	responseTopo, err := grpcclient.TopologyClient(&topoconf)

	if err != nil || responseTopo.ReturnCode == pb.ReturnCode_FAILED {
		topology.Status = entities.STATUS_FAILED
		database.Set(utils.KEY_PREFIX_TOPOLOGY+topology.Id, &topology)
		return fmt.Errorf("deploy topology failed, Error = '%s', return = '%s'", err.Error(), responseTopo.ReturnMessage)
	}

	topology.Status = entities.STATUS_READY
	database.Set(utils.KEY_PREFIX_TOPOLOGY+topology.Id, &topology)

	return nil
}

func NetworkHandler(s *entities.Scenario, responseTopo *pb.ReturnTopologyMessage, responseNetwork *pb.ReturnNetworkMessage) error {
	var network entities.NetworkConfig
	if err := database.FindEntity(s.NetworkConfId, utils.KEY_PREFIX_NETWORK, &network); err != nil {
		return errors.New("network config not found")
	}

	if network.Status != entities.STATUS_NONE {
		return fmt.Errorf("network config '%s' is '%s' now)", network.Id, network.Status)
	}

	if responseTopo.ReturnCode != pb.ReturnCode_OK {
		return fmt.Errorf("topology is not ready")
	}

	var service entities.ServiceConfig
	if err := database.FindEntity(s.ServiceConfId, utils.KEY_PREFIX_SERVICE, &service); err != nil {
		return errors.New("service config not found")
	}

	var compute entities.ComputeConfig
	if err := database.FindEntity(s.ComputeConfId, utils.KEY_PREFIX_COMPUTE, &compute); err != nil {
		return fmt.Errorf("compute config '%s' not found", s.ComputeConfId)
	}

	if compute.Status != entities.STATUS_NONE {
		return fmt.Errorf("nework config '%s' is '%s' now", s.ComputeConfId, compute.Status)
	}

	var netconf pb.InternalNetConfigInfo
	if err := constructNetConfMessage(&network, &service, responseTopo, &netconf); err != nil {
		return errors.New("netconfig protobuf message error")
	}

	network.Status = entities.STATUS_DEPLOYING
	database.Set(utils.KEY_PREFIX_NETWORK+network.Id, &network)

	responseNetwork, err := grpcclient.NetworkClient(&netconf)

	if err != nil || responseNetwork.ReturnCode == pb.ReturnCode_FAILED {
		network.Status = entities.STATUS_FAILED
		database.Set(utils.KEY_PREFIX_NETWORK+network.Id, &network)
		return fmt.Errorf("deploy network failed, Error = '%s', return = '%s'", err.Error(), responseNetwork.ReturnMessage)
	}

	network.Status = entities.STATUS_READY
	database.Set(utils.KEY_PREFIX_NETWORK+network.Id, &network)

	return nil
}

func ComputeHanlder(s *entities.Scenario, responseTopo *pb.ReturnTopologyMessage, responseNetwork *pb.ReturnNetworkMessage, responseCompute *pb.ReturnMessage) error {
	var compute entities.ComputeConfig
	if err := database.FindEntity(s.ComputeConfId, utils.KEY_PREFIX_COMPUTE, &compute); err != nil {
		return errors.New("compute config not found")
	}

	if compute.Status != entities.STATUS_NONE {
		return fmt.Errorf("compute config '%s' is '%s' now)", compute.Id, compute.Status)
	}

	if responseNetwork.ReturnCode != pb.ReturnCode_OK {
		return fmt.Errorf("network is not ready")
	}

	var service entities.ServiceConfig
	if err := database.FindEntity(s.ServiceConfId, utils.KEY_PREFIX_SERVICE, &service); err != nil {
		return errors.New("service config not found")
	}

	for _, n := range responseNetwork.Vpcs {
		for _, s := range n.Subnets {
			s.NumberVms = uint32(compute.NumberOfVmPerVpc) / uint32(compute.NumberOfComputeNodes) / uint32(len(n.Subnets))
		}
	}

	var computeconf pb.InternalComputeConfigInfo
	if err := constructComputeMessage(&compute, &service, responseTopo, responseNetwork, &computeconf); err != nil {
		return errors.New("compute protobuf message error")
	}

	compute.Status = entities.STATUS_DEPLOYING
	database.Set(utils.KEY_PREFIX_COMPUTE+compute.Id, &compute)

	responseCompute, err := grpcclient.ComputeClient(&computeconf)

	if err != nil || responseCompute.ReturnCode == pb.ReturnCode_FAILED {
		compute.Status = entities.STATUS_FAILED
		database.Set(utils.KEY_PREFIX_COMPUTE+compute.Id, &compute)
		return fmt.Errorf("deploy compute failed, Error = '%s', return = '%s'", err.Error(), responseCompute.ReturnMessage)
	}

	compute.Status = entities.STATUS_READY
	database.Set(utils.KEY_PREFIX_COMPUTE+compute.Id, &compute)

	return nil
}

func constructTopologyMessage(topo *entities.TopologyConfig, topoPb *pb.InternalTopologyInfo) error {
	topoPb.OperationType = pb.OperationType_CREATE
	topoPb.Config.FormatVersion = 1
	topoPb.Config.RevisionNumber = 1
	topoPb.Config.RequestId = utils.GenUUID()
	topoPb.Config.TopologyId = topo.Id
	topoPb.Config.Name = topo.Name
	topoPb.Config.MessageType = pb.MessageType_FULL
	topoPb.Config.TopologyType = getTopoloyType(topo.TopoType)
	topoPb.Config.NumberOfVhosts = uint32(topo.NumberOfVhosts)
	topoPb.Config.NumberOfRacks = uint32(topo.NumberOfRacks)
	topoPb.Config.VhostPerRack = uint32(topo.VhostsPerRack)
	topoPb.Config.DataPlaneCidr = topo.DataPlaneCidr
	topoPb.Config.NumberOfGateways = uint32(topo.NumberOfGateways)
	topoPb.Config.GatewayIps = topo.GatewayIPs

	for _, image := range topo.Images {
		var imagePb pb.InternalTopologyImage
		imagePb.OperationType = pb.OperationType_CREATE
		imagePb.Id = image.Id
		imagePb.Name = image.Name
		imagePb.Cmd = image.Cmd
		imagePb.Args = image.Args
		imagePb.Registry = image.Registry
		topoPb.Config.Images = append(topoPb.Config.Images, &imagePb)
	}

	for _, vnode := range topo.VNodes {
		var vnodePb pb.InternalVNodeInfo
		vnodePb.OperationType = pb.OperationType_CREATE
		vnodePb.Name = vnode.Name
		vnodePb.Type = getVNodeType(vnode.Type)
		for _, vnic := range vnode.Nics {
			var vnicPb pb.InternalVNicInfo
			vnicPb.Name = vnic.Name
			vnicPb.Ip = vnic.Ip
			vnodePb.Vnics = append(vnodePb.Vnics, &vnicPb)
		}
		topoPb.Config.Vnodes = append(topoPb.Config.Vnodes, &vnodePb)
	}

	return nil
}

func getTopoloyType(topoType string) pb.TopologyType {
	switch strings.ToLower(topoType) {
	case "linear":
		return pb.TopologyType_LINEAR
	case "single":
		return pb.TopologyType_SINGLE
	case "reversed":
		return pb.TopologyType_REVERSED
	case "mesh":
		return pb.TopologyType_MESH
	case "custom":
		return pb.TopologyType_CUSTOM
	default:
		return pb.TopologyType_TREE
	}
}

func getVNodeType(vnodeType string) pb.VNodeType {
	switch strings.ToLower(vnodeType) {
	case "vswitch":
		return pb.VNodeType_VSWITCH
	case "vrouter":
		return pb.VNodeType_VROUTER
	case "vgateway":
		return pb.VNodeType_VGATEWAY
	default:
		return pb.VNodeType_VHOST
	}
}

func constructNetConfMessage(netconf *entities.NetworkConfig, serviceConf *entities.ServiceConfig, topoReturn *pb.ReturnTopologyMessage, netconfPb *pb.InternalNetConfigInfo) error {
	netconfPb.OperationType = pb.OperationType_CREATE
	netconfPb.Config.FormatVersion = 1
	netconfPb.Config.RevisionNumber = 1
	netconfPb.Config.RequestId = utils.GenUUID()
	netconfPb.Config.NetconfigId = netconf.Id
	netconfPb.Config.MessageType = pb.MessageType_FULL

	var servicePb pb.InternalServiceInfo
	for _, service := range serviceConf.Services {
		if strings.ToUpper(service.WhereToRun) == utils.MERAK_NETWORK {
			servicePb.OperationType = pb.OperationType_CREATE
			servicePb.Id = service.Id
			servicePb.Name = service.Name
			servicePb.Cmd = service.Cmd
			servicePb.Url = service.Url
			servicePb.Parameters = service.Parameters
			servicePb.ReturnCode = service.ReturnCode
			servicePb.ReturnString = service.ReturnString
			servicePb.WhenToRun = service.WhenToRun
			servicePb.WhereToRun = service.WhereToRun
			netconfPb.Config.Services = append(netconfPb.Config.Services, &servicePb)
		}
	}

	netconfPb.Config.Network.OperationType = pb.OperationType_CREATE
	netconfPb.Config.Network.Id = netconf.Id
	netconfPb.Config.Network.Name = netconf.Name
	netconfPb.Config.Network.NumberOfVpcs = uint32(netconf.NumberOfVPCS)
	netconfPb.Config.Network.NumberOfSubnetPerVpc = uint32(netconf.NumberOfSubnetPerVpc)
	netconfPb.Config.Network.NumberOfSecurityGroups = uint32(netconf.NumberOfSecurityGroups)

	var vpcPb pb.InternalVpcInfo
	for _, vpc := range netconf.Vpcs {
		vpcPb.VpcId = vpc.VpcId
		vpcPb.TenantId = vpc.TenantId
		vpcPb.ProjectId = vpc.ProjectId

		var subnetPb pb.InternalSubnetInfo
		for _, subnet := range vpc.SubnetInfo {
			subnetPb.SubnetId = subnet.SubnetId
			subnetPb.SubnetCidr = subnet.SubnetCidr
			subnetPb.SubnetGw = subnet.SubnetGateway
			vpcPb.Subnets = append(vpcPb.Subnets, &subnetPb)
		}
		netconfPb.Config.Network.Vpcs = append(netconfPb.Config.Network.Vpcs, &vpcPb)
	}

	var routerPb pb.InternalRouterInfo
	for _, router := range netconf.Routers {
		routerPb.OperationType = pb.OperationType_CREATE
		routerPb.Id = router.Id
		routerPb.Name = router.Name
		routerPb.Subnets = router.SubnetGateways
		netconfPb.Config.Network.Routers = append(netconfPb.Config.Network.Routers, &routerPb)
	}

	var gatewayPb pb.InternalGatewayInfo
	for _, gateway := range netconf.Gateways {
		gatewayPb.OperationType = pb.OperationType_CREATE
		gatewayPb.Id = gateway.Id
		gatewayPb.Name = gateway.Name
		gatewayPb.Ips = gateway.Ips
		netconfPb.Config.Network.Gateways = append(netconfPb.Config.Network.Gateways, &gatewayPb)
	}

	var sgPb pb.InternalSecurityGroupInfo
	for _, sg := range netconf.SecurityGroups {
		sgPb.OperationType = pb.OperationType_CREATE
		sgPb.Id = sg.Id
		sgPb.Name = sg.Name
		sgPb.ApplyTo = sg.ApplyTo

		var sgRulePb pb.InternalSecurityGroupRulelnfo
		for _, rule := range sg.Rules {
			sgRulePb.OperationType = pb.OperationType_CREATE
			sgRulePb.Id = rule.Id
			sgRulePb.Name = rule.Name
			sgRulePb.Description = rule.Description
			sgRulePb.Ethertype = rule.EtherType
			sgRulePb.Protocol = rule.Protocol
			sgRulePb.PortRange = rule.PortRange
			sgRulePb.RemoteGroupId = rule.RemoteGroupId
			sgRulePb.RemoteIpPrefix = rule.RemoteIpPrefix
			sgPb.Rules = append(sgPb.Rules, &sgRulePb)
		}
		netconfPb.Config.Network.SecurityGroups = append(netconfPb.Config.Network.SecurityGroups, &sgPb)
	}

	netconfPb.Config.Computes = topoReturn.ComputeNodes

	return nil
}

func constructComputeMessage(compute *entities.ComputeConfig, serviceConf *entities.ServiceConfig, topoReturn *pb.ReturnTopologyMessage, netReturn *pb.ReturnNetworkMessage, computePb *pb.InternalComputeConfigInfo) error {
	computePb.OperationType = pb.OperationType_CREATE
	computePb.Config.FormatVersion = 1
	computePb.Config.RevisionNumber = 1
	computePb.Config.RequestId = utils.GenUUID()
	computePb.Config.ComputeConfigId = compute.Id
	computePb.Config.MessageType = pb.MessageType_FULL
	computePb.Config.Pods = topoReturn.ComputeNodes
	computePb.Config.VmDeploy.OperationType = pb.OperationType_CREATE
	computePb.Config.VmDeploy.Vpcs = netReturn.Vpcs
	computePb.Config.VmDeploy.Secgroups = netReturn.SecurityGroupIds
	computePb.Config.VmDeploy.DeployType = getVMDeployType(compute.VmDeployType)
	computePb.Config.VmDeploy.Scheduler = getVMDeployScheduler(compute.Scheduler)

	var servicePb pb.InternalServiceInfo
	for _, service := range serviceConf.Services {
		if strings.ToUpper(service.WhereToRun) == utils.MERAK_COMPUTE || strings.ToUpper(service.WhereToRun) == utils.MERAK_AGENT {
			servicePb.OperationType = pb.OperationType_CREATE
			servicePb.Id = service.Id
			servicePb.Name = service.Name
			servicePb.Cmd = service.Cmd
			servicePb.Url = service.Url
			servicePb.Parameters = service.Parameters
			servicePb.ReturnCode = service.ReturnCode
			servicePb.ReturnString = service.ReturnString
			servicePb.WhenToRun = service.WhenToRun
			servicePb.WhereToRun = service.WhereToRun
			computePb.Config.Services = append(computePb.Config.Services, &servicePb)
		}
	}

	return nil
}

func getVMDeployType(deploy string) pb.VMDeployType {
	switch strings.ToLower(deploy) {
	case "assign":
		return pb.VMDeployType_ASSIGN
	case "skew":
		return pb.VMDeployType_SKEW
	case "random":
		return pb.VMDeployType_RANDOM
	default:
		return pb.VMDeployType_UNIFORM
	}
}

func getVMDeployScheduler(scheduler string) pb.VMScheduleType {
	switch strings.ToLower(scheduler) {
	case "sequential":
		return pb.VMScheduleType_SEQUENTIAL
	case "skew":
		return pb.VMScheduleType_RPS
	case "random":
		return pb.VMScheduleType_RANDOM_SCHEDULE
	default:
		return pb.VMScheduleType_SEQUENTIAL
	}
}
