package activities

import (
	"context"
	"encoding/json"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-network/entities"
	"github.com/futurewei-cloud/merak/services/merak-network/http"
	"log"
)

func doVPC(vpc *pb.InternalVpcInfo) (vpcId string) {
	log.Println("doVPC")
	vpcBody := entities.VpcStruct{Network: entities.VpcBody{
		AdminStateUp:        true,
		RevisionNumber:      0,
		Cidr:                vpc.VpcCidr,
		ByDefault:           true,
		Description:         "vpc",
		DnsDomain:           "domain",
		IsDefault:           true,
		Mtu:                 1400,
		Name:                "YM_sample_vpc",
		PortSecurityEnabled: true,
		ProjectId:           vpc.ProjectId,
	}}
	returnMessage, returnErr := http.RequestCall("http://54.188.252.43:30001/project/123456789/vpcs", "POST", vpcBody)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.VpcReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("doVPC done")
	return returnJson.Network.ID
}
func doSubnet(subnet *pb.InternalSubnetInfo, vpcId string) (subnetId string) {
	log.Println("doSubnet")
	subnetBody := entities.SubnetStruct{Subnet: entities.SubnetBody{
		Cider:     subnet.SubnetCidr,
		Name:      "YM_sample_subnet",
		IpVersion: 4,
		NetworkId: vpcId,
	}}
	returnMessage, returnErr := http.RequestCall("http://54.188.252.43:30002/project/123456789/subnets", "POST", subnetBody)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.SubnetReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("doVPC done")
	return returnJson.Subnet.ID
}
func doRouter(vpcId string) (routerId string) {
	log.Println("doRouter")
	routerBody := entities.RouterStruct{Router: entities.RouterBody{
		AdminStateUp: true,
		Description:  "router description",
		Distributed:  true,
		ExternalGatewayInfo: entities.RouterExternalGatewayInfo{
			EnableSnat:       true,
			ExternalFixedIps: nil,
			NetworkId:        vpcId,
		},
		FlavorId:       "",
		GatewayPorts:   nil,
		Ha:             true,
		Name:           "YM_simple_router",
		ProjectId:      "123456789",
		RevisionNumber: 0,
		Status:         "BUILD",
		TenantId:       "123456789",
	}}
	returnMessage, returnErr := http.RequestCall("http://54.188.252.43:30003/project/123456789/routers", "POST", routerBody)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.RouterReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("doRouter done")
	return returnJson.Router.ID
}

func doAttachRouter(routerId string, subnetId string) error {
	log.Println("doAttachRouter")
	attachRouterBody := entities.AttachRouterStruct{SubnetId: subnetId}
	url := "http://54.188.252.43:30003/project/123456789/routers/" + routerId + "/add_router_interface"
	returnMessage, returnErr := http.RequestCall(url, "PUT", attachRouterBody)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.AttachRouterReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("doAttachRouter done")
	return nil
}

func VnetCreate(ctx context.Context, network *pb.InternalNetworkInfo) (string, error) {
	log.Println("VnetCreate")
	// TODO may want to separate bellow sections to different function, and use `go` and `wg` to improve overall speed
	// TODO when do concurrent, need to keep in mind on how to control the number of concurrency
	// Doing vpc and subnet
	var returnInfo []*pb.InternalVpcInfo
	var vpcId string
	subnetCiderIdMap := make(map[string]string)
	for i := 0; i < int(network.NumberOfVpcs); i++ {
		vpcId = doVPC(network.Vpcs[i])

		var subnetInfo []*pb.InternalSubnetInfo
		for j := 0; j < int(network.NumberOfSubnetPerVpc); j++ {
			//subnetId := utils.GenUUID()
			subnetId := doSubnet(network.Vpcs[i].Subnets[j], vpcId)
			subnetCiderIdMap[network.Vpcs[i].Subnets[j].SubnetCidr] = subnetId
			log.Printf("subnetCiderIdMap %s", subnetCiderIdMap)
			currentSubnet := pb.InternalSubnetInfo{
				SubnetId:   subnetId,
				SubnetCidr: network.Vpcs[i].Subnets[j].SubnetCidr,
				SubnetGw:   network.Vpcs[i].Subnets[j].SubnetGw,
				NumberVms:  network.Vpcs[i].Subnets[j].NumberVms,
			}
			subnetInfo = append(subnetInfo, &currentSubnet)
		}
		currentVPC := pb.InternalVpcInfo{
			VpcId:     vpcId,
			TenantId:  network.Vpcs[i].TenantId,
			ProjectId: network.Vpcs[i].ProjectId,
			Subnets:   subnetInfo,
		}
		returnInfo = append(returnInfo, &currentVPC)
		log.Printf("VnetCreate End %s", returnInfo)

	}

	//doing router: create and attach subnet
	for _, router := range network.Routers {
		routerId := doRouter(vpcId)
		for _, subnet := range router.Subnets {
			doAttachRouter(routerId, subnetCiderIdMap[subnet])
		}
	}
	return "VnetCreate", nil
}
