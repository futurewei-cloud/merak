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
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	compute_pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	network_pb "github.com/futurewei-cloud/merak/api/proto/v1/network"
	topology_pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	"github.com/futurewei-cloud/merak/services/scenario-manager/utils"
)

func constructTopologyMessage(topo *entities.TopologyConfig, serviceConf *entities.ServiceConfig, topoPb *topology_pb.InternalTopologyInfo, action entities.EventName) error {
	topoPb.OperationType = actionToOperation(action)
	var conf topology_pb.InternalTopologyConfiguration
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
	conf.PortsPerVswitch = uint32(topo.PortsPerVSwitch)
	conf.DataPlaneCidr = topo.DataPlaneCidr
	conf.NumberOfGateways = uint32(topo.NumberOfGateways)
	conf.GatewayIps = topo.GatewayIPs

	if serviceConf != nil {
		for _, service := range serviceConf.Services {
			var servicePb pb.InternalServiceInfo
			if strings.ToUpper(service.WhereToRun) == utils.MERAK_AGENT {
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

	for _, image := range topo.Images {
		var imagePb topology_pb.InternalTopologyImage
		imagePb.OperationType = actionToOperation(action)
		imagePb.Id = image.Id
		imagePb.Name = image.Name
		imagePb.Cmd = image.Cmd
		imagePb.Args = image.Args
		imagePb.Registry = image.Registry
		conf.Images = append(conf.Images, &imagePb)
	}

	for _, vnode := range topo.VNodes {
		var vnodePb topology_pb.InternalVNodeInfo
		vnodePb.OperationType = actionToOperation(action)
		vnodePb.Name = vnode.Name
		vnodePb.Type = getVNodeType(vnode.Type)
		for _, vnic := range vnode.Nics {
			var vnicPb topology_pb.InternalVNicInfo
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

func getTopoloyType(topoType string) topology_pb.TopologyType {
	switch strings.ToLower(topoType) {
	case "linear":
		return topology_pb.TopologyType_LINEAR
	case "single":
		return topology_pb.TopologyType_SINGLE
	case "reversed":
		return topology_pb.TopologyType_REVERSED
	case "mesh":
		return topology_pb.TopologyType_MESH
	case "custom":
		return topology_pb.TopologyType_CUSTOM
	default:
		return topology_pb.TopologyType_TREE
	}
}

func getVNodeType(vnodeType string) topology_pb.VNodeType {
	switch strings.ToLower(vnodeType) {
	case "vswitch":
		return topology_pb.VNodeType_VSWITCH
	case "vrouter":
		return topology_pb.VNodeType_VROUTER
	case "vgateway":
		return topology_pb.VNodeType_VGATEWAY
	default:
		return topology_pb.VNodeType_VHOST
	}
}

func constructNetConfMessage(netconf *entities.NetworkConfig, serviceConf *entities.ServiceConfig, topoReturn *topology_pb.ReturnTopologyMessage, netconfPb *network_pb.InternalNetConfigInfo, action entities.EventName) error {
	netconfPb.OperationType = actionToOperation(action)
	var conf network_pb.InternalNetConfigConfiguration
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

	var netPb network_pb.InternalNetworkInfo
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
		var routerPb network_pb.InternalRouterInfo
		routerPb.OperationType = actionToOperation(action)
		routerPb.Id = router.Id
		routerPb.Name = router.Name
		routerPb.Subnets = router.SubnetGateways
		netPb.Routers = append(netPb.Routers, &routerPb)
	}

	for _, gateway := range netconf.Gateways {
		var gatewayPb network_pb.InternalGatewayInfo
		gatewayPb.OperationType = actionToOperation(action)
		gatewayPb.Id = gateway.Id
		gatewayPb.Name = gateway.Name
		gatewayPb.Ips = gateway.Ips
		netPb.Gateways = append(netPb.Gateways, &gatewayPb)
	}

	for _, sg := range netconf.SecurityGroups {
		var sgPb network_pb.InternalSecurityGroupInfo
		sgPb.OperationType = actionToOperation(action)
		sgPb.Id = sg.Id
		sgPb.Name = sg.Name
		sgPb.ApplyTo = sg.ApplyTo

		for _, rule := range sg.Rules {
			var sgRulePb network_pb.InternalSecurityGroupRulelnfo
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
	if topoReturn != nil {
		conf.Computes = topoReturn.GetComputeNodes()
	}
	netconfPb.Config = &conf

	return nil
}

func constructComputeMessage(compute *entities.ComputeConfig, serviceConf *entities.ServiceConfig, topoReturn *topology_pb.ReturnTopologyMessage, netReturn *network_pb.ReturnNetworkMessage, computePb *compute_pb.InternalComputeConfigInfo, action entities.EventName) error {
	computePb.OperationType = actionToOperation(action)

	var conf compute_pb.InternalComputeConfiguration
	conf.FormatVersion = 1
	conf.RevisionNumber = 1
	conf.RequestId = utils.GenUUID()
	conf.ComputeConfigId = compute.Id
	conf.MessageType = pb.MessageType_FULL
	if topoReturn != nil {
		conf.Pods = topoReturn.GetComputeNodes()
	}

	var vmDeployPb compute_pb.InternalVMDeployInfo
	vmDeployPb.OperationType = actionToOperation(action)
	vmDeployPb.DeployType = getVMDeployType(compute.VmDeployType)
	vmDeployPb.Scheduler = getVMDeployScheduler(compute.Scheduler)

	if action != entities.EVENT_CHECK {
		if netReturn != nil {
			vmDeployPb.Secgroups = netReturn.GetSecurityGroupIds()
		} else {
			return errors.New("construct compute message - virtual network is not ready yet")
		}
		if vmDeployPb.DeployType == compute_pb.VMDeployType_ASSIGN {
			if len(compute.VPCInfo) <= 0 {
				return errors.New("construct compute message - please enter VPCInfo for creating VM")
			}
			for _, vpc := range compute.VPCInfo {
				var vpcPb pb.InternalVpcInfo
				vpcPb.ProjectId = vpc.ProjectId
				vpcPb.TenantId = vpc.TenantId
				vpcPb.VpcCidr = vpc.VpcCidr
				if len(vpc.SubnetInfo) <= 0 {
					return errors.New("construct compute message - please enter subnet_info and number of VMs to be deployed")
				}
				for _, subnet := range vpc.SubnetInfo {
					var subnetPb pb.InternalSubnetInfo
					subnetPb.NumberVms = uint32(subnet.NumberOfVMs)
					subnetPb.SubnetCidr = subnet.SubnetCidr
					subnetPb.SubnetGw = subnet.SubnetGateway
					vpcPb.Subnets = append(vpcPb.Subnets, &subnetPb)
				}
				vmDeployPb.Vpcs = append(vmDeployPb.Vpcs, &vpcPb)
			}
		} else {
			vmDeployPb.Vpcs = netReturn.GetVpcs()
			for _, vpc := range vmDeployPb.GetVpcs() {
				for _, subnet := range vpc.GetSubnets() {
					if compute.NumberOfComputeNodes != 0 && len(vpc.GetSubnets()) != 0 {
						subnet.NumberVms = uint32(compute.NumberOfVmPerVpc) / uint32(compute.NumberOfComputeNodes) / uint32(len(vpc.GetSubnets()))
					}
					if subnet.NumberVms <= 0 {
						return errors.New("construct compute message - number of VMs to be deployed in a VPC are zero")
					}
				}
			}
		}
	}

	conf.VmDeploy = &vmDeployPb

	if serviceConf != nil {
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
	}
	computePb.Config = &conf

	return nil
}

func getVMDeployType(deploy string) compute_pb.VMDeployType {
	switch strings.ToLower(deploy) {
	case "assign":
		return compute_pb.VMDeployType_ASSIGN
	case "skew":
		return compute_pb.VMDeployType_SKEW
	case "random":
		return compute_pb.VMDeployType_RANDOM
	default:
		return compute_pb.VMDeployType_UNIFORM
	}
}

func getVMDeployScheduler(scheduler string) compute_pb.VMScheduleType {
	switch strings.ToLower(scheduler) {
	case "sequential":
		return compute_pb.VMScheduleType_SEQUENTIAL
	case "skew":
		return compute_pb.VMScheduleType_RPS
	case "random":
		return compute_pb.VMScheduleType_RANDOM_SCHEDULE
	default:
		return compute_pb.VMScheduleType_SEQUENTIAL
	}
}
