package activities

import (
	"context"
	"encoding/json"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-network/database"
	"github.com/futurewei-cloud/merak/services/merak-network/entities"
	"github.com/futurewei-cloud/merak/services/merak-network/http"
	"github.com/futurewei-cloud/merak/services/merak-network/utils"
	"log"
	"sync"
)

func getSubnetRouter(subnetId string, projectId string) (returnRouterId string) {
	log.Println("getSubnetRouter")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30002/project/"+projectId+"/subnets/"+subnetId, "GET", "", nil)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.SubnetReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("returnJson : %s", returnJson)
	log.Println("getSubnetRouter done")
	return returnJson.Subnet.AttachedRouterID
}

func removeInterfaceToNeutronRouter(subnetId string, routerId string, projectId string) (returnRouterId string) {
	log.Println("removeInterfaceToNeutronRouter")
	payloadBody := entities.RemoveInterfaceToNeutronRouterBody{
		SubnetId: subnetId,
	}
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30003/project/"+projectId+"/routers/"+routerId+"/remove_router_interface", "PUT", payloadBody, nil)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.RemoveInterfaceToNeutronRouterReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("removeInterfaceToNeutronRouter returnJson : %s", returnJson)
	log.Println("removeInterfaceToNeutronRouter done")

	return returnJson.ID
}

func deleteNeutronRouterByRouterId(routerId string, projectId string) (returnRouterId string) {
	log.Println("deleteNeutronRouterByRouterId")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30003/project/"+projectId+"/routers/"+routerId, "DELETE", "", nil)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	database.Del(utils.Router + routerId)
	var returnJson entities.DeleteNeutronRouterByRouterIdReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("deleteNeutronRouterByRouterId returnJson : %s", returnJson)
	log.Println("deleteNeutronRouterByRouterId done")
	return returnJson.ID
}
func deleteSubnet(subnetId string, projectId string) (returnSubnetId string) {
	log.Println("deleteSubnet")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30002/project/"+projectId+"/subnets/"+subnetId, "DELETE", "", nil)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	database.Del(utils.SUBNET + subnetId)
	log.Println("deleteSubnet done")
	return subnetId
}

func deleteVpc(vpcId string, projectId string) (returnVpcId string) {
	log.Println("deleteVpc")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30001/project/"+projectId+"/vpcs/"+vpcId, "DELETE", "", nil)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	database.Del(utils.VPC + vpcId)
	log.Println("deleteVpc done")
	return vpcId
}

func deleteSg(sgId string, projectId string) (returnVpcId string) {
	log.Println("deleteSg")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30001/project/"+projectId+"/security-groups/"+sgId, "DELETE", "", nil)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	database.Del(utils.SECURITYGROUP + sgId)
	log.Println("deleteVpc done")
	return sgId
}

func VnetDelete(ctx context.Context, netConfigId string, wg *sync.WaitGroup, returnMessage chan *pb.ReturnNetworkMessage) (string, error) {
	defer wg.Done()
	// TODO: when query db, make sure to check if key exist first, other wise could timeout
	log.Println("VnetDelete")
	values, err := database.Get(utils.NETCONFIG + netConfigId)
	if err != nil {
		return "", err
	}
	log.Printf("VnetInfo %s", values)
	log.Printf("returnMessage %s", returnMessage)
	var returnJson *pb.ReturnNetworkMessage
	json.Unmarshal([]byte(values), &returnJson)
	log.Printf("returnMessage %s", returnJson)

	for _, vpc := range returnJson.Vpcs {
		projectId := vpc.ProjectId
		var routerIds []string
		routerIdsMap := make(map[string]int) //to keep track if the router already been appended to routerIds
		for _, subnet := range vpc.Subnets {
			routerId := getSubnetRouter(subnet.SubnetId, projectId)
			if routerId != "" {
				returnRouterId := removeInterfaceToNeutronRouter(subnet.SubnetId, routerId, projectId)
				if returnRouterId != routerId {
					log.Fatalf("Router Subnet Interface delete fail %s", routerId)
				}
				if _, ok := routerIdsMap[routerId]; !ok {
					//if routerId have not been appened to the routerIds yet
					routerIdsMap[routerId] = 1
					routerIds = append(routerIds, routerId)
				}
			}
			//deleteRouteTable(subnet.SubnetId, projectId)
			deleteSubnet(subnet.SubnetId, projectId)
		}
		//deleteRouter(vpc.VpcId, projectId)
		for _, routerId := range routerIds {
			returnRouterId := deleteNeutronRouterByRouterId(routerId, projectId)
			if returnRouterId != routerId {
				log.Fatalf("Router delete fail %s", routerId)
			}
		}
		deleteVpc(vpc.VpcId, projectId)
	}
	for _, sgId := range returnJson.SecurityGroupIds {
		//TODO: need find new way to get projectId
		projectId := returnJson.Vpcs[0].ProjectId
		deleteSg(sgId, projectId)
	}
	database.Del(utils.NETCONFIG + netConfigId)

	json.Unmarshal([]byte(values), &returnJson)
	log.Printf("returnMessage %s", returnJson)

	returnMessage <- returnJson
	return "VnetInfo", nil
}
