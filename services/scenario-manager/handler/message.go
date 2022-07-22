package handler

import (
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	"github.com/futurewei-cloud/merak/services/scenario-manager/utils"
)

func constructTopologyMessage(topo *entities.TopologyConfig, topoPb *pb.InternalTopologyInfo, action entities.EventName) error {
	topoPb.OperationType = actionToOperation(action)
	var conf pb.InternalTopologyConfiguration
	conf.FormatVersion = 1
	conf.RevisionNumber = 1
	conf.RequestId = utils.GenUUID()
	conf.TopologyId = topo.Id
	conf.Name = topo.Name
	conf.MessageType = pb.MessageType_FULL
	conf.TopologyType = getTopoloyType(topo.TopoType)
	conf.NumberOfVhosts = uint32(topo.NumberOfVhosts)
	conf.NumberOfRacks = uint32(topo.NumberOfRacks)
	conf.VhostPerRack = uint32(topo.VhostsPerRack)
	conf.DataPlaneCidr = topo.DataPlaneCidr
	conf.NumberOfGateways = uint32(topo.NumberOfGateways)
	conf.GatewayIps = topo.GatewayIPs

	for _, image := range topo.Images {
		var imagePb pb.InternalTopologyImage
		imagePb.OperationType = actionToOperation(action)
		imagePb.Id = image.Id
		imagePb.Name = image.Name
		imagePb.Cmd = image.Cmd
		imagePb.Args = image.Args
		imagePb.Registry = image.Registry
		conf.Images = append(conf.Images, &imagePb)
	}

	for _, vnode := range topo.VNodes {
		var vnodePb pb.InternalVNodeInfo
		vnodePb.OperationType = actionToOperation(action)
		vnodePb.Name = vnode.Name
		vnodePb.Type = getVNodeType(vnode.Type)
		for _, vnic := range vnode.Nics {
			var vnicPb pb.InternalVNicInfo
			vnicPb.Name = vnic.Name
			vnicPb.Ip = vnic.Ip
			vnodePb.Vnics = append(vnodePb.Vnics, &vnicPb)
		}
		conf.Vnodes = append(conf.Vnodes, &vnodePb)
	}

	topoPb.Config = &conf

	return nil
}

func actionToOperation(action entities.EventName) pb.OperationType {
	switch action {
	case entities.EVENT_DEPLOY:
		return pb.OperationType_CREATE
	case entities.EVENT_DELETE:
		return pb.OperationType_DELETE
	case entities.EVENT_UPDATE:
		return pb.OperationType_UPDATE
	default:
		return pb.OperationType_INFO
	}
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

func constructNetConfMessage(netconf *entities.NetworkConfig, serviceConf *entities.ServiceConfig, topoReturn *pb.ReturnTopologyMessage, netconfPb *pb.InternalNetConfigInfo, action entities.EventName) error {
	netconfPb.OperationType = actionToOperation(action)
	var conf pb.InternalNetConfigConfiguration
	conf.FormatVersion = 1
	conf.RevisionNumber = 1
	conf.RequestId = utils.GenUUID()
	conf.NetconfigId = netconf.Id
	conf.MessageType = pb.MessageType_FULL

	if serviceConf != nil {
		for _, service := range serviceConf.Services {
			var servicePb pb.InternalServiceInfo
			if strings.ToUpper(service.WhereToRun) == utils.MERAK_NETWORK {
				servicePb.OperationType = actionToOperation(action)
				servicePb.Id = service.Id
				servicePb.Name = service.Name
				servicePb.Cmd = service.Cmd
				servicePb.Url = service.Url
				servicePb.Parameters = service.Parameters
				servicePb.ReturnCode = service.ReturnCode
				servicePb.ReturnString = service.ReturnString
				servicePb.WhenToRun = service.WhenToRun
				servicePb.WhereToRun = service.WhereToRun
				conf.Services = append(conf.Services, &servicePb)
			}
		}
	}

	var netPb pb.InternalNetworkInfo
	netPb.OperationType = actionToOperation(action)
	netPb.Id = netconf.Id
	netPb.Name = netconf.Name
	netPb.NumberOfVpcs = uint32(netconf.NumberOfVPCS)
	netPb.NumberOfSubnetPerVpc = uint32(netconf.NumberOfSubnetPerVpc)
	netPb.NumberOfSecurityGroups = uint32(netconf.NumberOfSecurityGroups)

	for _, vpc := range netconf.Vpcs {
		var vpcPb pb.InternalVpcInfo
		vpcPb.VpcId = vpc.VpcId
		vpcPb.TenantId = vpc.TenantId
		vpcPb.ProjectId = vpc.ProjectId
		vpcPb.VpcCidr = vpc.VpcCidr

		for _, subnet := range vpc.SubnetInfo {
			var subnetPb pb.InternalSubnetInfo
			subnetPb.SubnetId = subnet.SubnetId
			subnetPb.SubnetCidr = subnet.SubnetCidr
			subnetPb.SubnetGw = subnet.SubnetGateway
			subnetPb.NumberVms = uint32(subnet.NumberOfVMs)
			vpcPb.Subnets = append(vpcPb.Subnets, &subnetPb)
		}
		netPb.Vpcs = append(netPb.Vpcs, &vpcPb)
	}

	for _, router := range netconf.Routers {
		var routerPb pb.InternalRouterInfo
		routerPb.OperationType = actionToOperation(action)
		routerPb.Id = router.Id
		routerPb.Name = router.Name
		routerPb.Subnets = router.SubnetGateways
		netPb.Routers = append(netPb.Routers, &routerPb)
	}

	for _, gateway := range netconf.Gateways {
		var gatewayPb pb.InternalGatewayInfo
		gatewayPb.OperationType = actionToOperation(action)
		gatewayPb.Id = gateway.Id
		gatewayPb.Name = gateway.Name
		gatewayPb.Ips = gateway.Ips
		netPb.Gateways = append(netPb.Gateways, &gatewayPb)
	}

	for _, sg := range netconf.SecurityGroups {
		var sgPb pb.InternalSecurityGroupInfo
		sgPb.OperationType = actionToOperation(action)
		sgPb.Id = sg.Id
		sgPb.Name = sg.Name
		sgPb.ApplyTo = sg.ApplyTo

		for _, rule := range sg.Rules {
			var sgRulePb pb.InternalSecurityGroupRulelnfo
			sgRulePb.OperationType = actionToOperation(action)
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
		netPb.SecurityGroups = append(netPb.SecurityGroups, &sgPb)
	}

	conf.Network = &netPb
	conf.Computes = topoReturn.GetComputeNodes()
	netconfPb.Config = &conf

	return nil
}

func constructComputeMessage(compute *entities.ComputeConfig, serviceConf *entities.ServiceConfig, topoReturn *pb.ReturnTopologyMessage, netReturn *pb.ReturnNetworkMessage, computePb *pb.InternalComputeConfigInfo, action entities.EventName) error {
	computePb.OperationType = pb.OperationType_CREATE

	var conf pb.InternalComputeConfiguration
	conf.FormatVersion = 1
	conf.RevisionNumber = 1
	conf.RequestId = utils.GenUUID()
	conf.ComputeConfigId = compute.Id
	conf.MessageType = pb.MessageType_FULL
	conf.Pods = topoReturn.GetComputeNodes()

	var vmDeployPb pb.InternalVMDeployInfo
	vmDeployPb.OperationType = pb.OperationType_CREATE
	vmDeployPb.Vpcs = netReturn.GetVpcs()
	vmDeployPb.Secgroups = netReturn.GetSecurityGroupIds()
	vmDeployPb.DeployType = getVMDeployType(compute.VmDeployType)
	vmDeployPb.Scheduler = getVMDeployScheduler(compute.Scheduler)

	conf.VmDeploy = &vmDeployPb

	for _, service := range serviceConf.Services {
		var servicePb pb.InternalServiceInfo
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
			conf.Services = append(conf.Services, &servicePb)
		}
	}

	computePb.Config = &conf

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
