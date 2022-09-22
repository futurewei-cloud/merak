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

package activities

import (
	"encoding/json"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/network"
	"github.com/futurewei-cloud/merak/services/merak-network/database"
	"github.com/futurewei-cloud/merak/services/merak-network/entities"
	"github.com/futurewei-cloud/merak/services/merak-network/http"
	"github.com/futurewei-cloud/merak/services/merak-network/utils"
	"log"
)

func getSubnetRouter(subnetId string, projectId string) (returnRouterId string, err error) {
	log.Println("getSubnetRouter")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30002/project/"+projectId+"/subnets/"+subnetId, "GET", "", nil)
	if returnErr != nil {
		log.Printf("returnErr %s", returnErr)
		return "", returnErr
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.SubnetReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("returnJson : %s", returnJson)
	log.Println("getSubnetRouter done")
	return returnJson.Subnet.AttachedRouterID, nil
}

func removeInterfaceToNeutronRouter(subnetId string, routerId string, projectId string) (returnRouterId string, err error) {
	log.Println("removeInterfaceToNeutronRouter")
	payloadBody := entities.RemoveInterfaceToNeutronRouterBody{
		SubnetId: subnetId,
	}
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30003/project/"+projectId+"/routers/"+routerId+"/remove_router_interface", "PUT", payloadBody, nil)
	if returnErr != nil {
		log.Printf("returnErr %s", returnErr)
		return "", returnErr
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.RemoveInterfaceToNeutronRouterReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("removeInterfaceToNeutronRouter returnJson : %s", returnJson)
	log.Println("removeInterfaceToNeutronRouter done")

	return returnJson.ID, nil
}

func deleteNeutronRouterByRouterId(routerId string, projectId string) (returnRouterId string, err error) {
	log.Println("deleteNeutronRouterByRouterId")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30003/project/"+projectId+"/routers/"+routerId, "DELETE", "", nil)
	if returnErr != nil {
		log.Printf("returnErr %s", returnErr)
		return "", returnErr
	}
	log.Printf("returnMessage %s", returnMessage)
	database.Del(utils.Router + routerId)
	var returnJson entities.DeleteNeutronRouterByRouterIdReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("deleteNeutronRouterByRouterId returnJson : %s", returnJson)
	log.Println("deleteNeutronRouterByRouterId done")
	return returnJson.ID, nil
}
func deleteSubnet(subnetId string, projectId string) (returnSubnetId string, err error) {
	log.Println("deleteSubnet")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30002/project/"+projectId+"/subnets/"+subnetId, "DELETE", "", nil)
	if returnErr != nil {
		log.Printf("returnErr %s", returnErr)
		return "", returnErr
	}
	log.Printf("returnMessage %s", returnMessage)
	database.Del(utils.SUBNET + subnetId)
	log.Println("deleteSubnet done")
	return subnetId, nil
}

func deleteVpc(vpcId string, projectId string) (returnVpcId string, err error) {
	log.Println("deleteVpc")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30001/project/"+projectId+"/vpcs/"+vpcId, "DELETE", "", nil)
	if returnErr != nil {
		log.Printf("returnErr %s", returnErr)
		return "", returnErr
	}
	log.Printf("returnMessage %s", returnMessage)
	database.Del(utils.VPC + vpcId)
	log.Println("deleteVpc done")
	return vpcId, nil
}

func deleteSg(sgId string, projectId string) (returnSgId string, err error) {
	log.Println("deleteSg")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30008/project/"+projectId+"/security-groups/"+sgId, "DELETE", "", nil)
	if returnErr != nil {
		log.Printf("returnErr %s", returnErr)
		return "", returnErr
	}
	log.Printf("returnMessage %s", returnMessage)
	database.Del(utils.SECURITYGROUP + sgId)
	log.Println("deleteVpc done")
	return sgId, nil
}

func deleteNode(netConfigId string) (err error) {
	log.Println("deleteNode")
	values, err := database.Get(utils.NODEGROUP + netConfigId)
	if err != nil {
		return err
	}
	log.Printf("NodeGroup %s", values)
	var returnJson entities.NodeReturn
	json.Unmarshal([]byte(values), &returnJson)
	log.Printf("returnMessage %s", returnJson)

	_, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30007/nodes/bulk", "DELETE", nil, nil)
	if returnErr != nil {
		log.Printf("returnErr %s", returnErr)
		return returnErr
	}

	database.Del(utils.NODEGROUP + netConfigId)
	log.Println("deleteNode done")
	return nil
}

func VnetDelete(netConfigId string, returnMessage chan *pb.ReturnNetworkMessage) (*pb.ReturnNetworkMessage, error) {
	// TODO: when query db, make sure to check if key exist first, other wise could timeout
	log.Println("VnetDelete")

	deleteNode(netConfigId)

	values, err := database.Get(utils.NETCONFIG + netConfigId)
	if err != nil {
		return nil, err
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
			routerId, err := getSubnetRouter(subnet.SubnetId, projectId)
			if err != nil {
				return nil, err
			}
			if routerId != "" {
				returnRouterId, err := removeInterfaceToNeutronRouter(subnet.SubnetId, routerId, projectId)
				if err != nil {
					return nil, err
				}
				if returnRouterId != routerId {
					log.Printf("Router Subnet Interface delete fail %s", routerId)
				}
				if _, ok := routerIdsMap[routerId]; !ok {
					//if routerId have not been appened to the routerIds yet
					routerIdsMap[routerId] = 1
					routerIds = append(routerIds, routerId)
				}
			}
			deleteSubnet(subnet.SubnetId, projectId)
		}
		for _, routerId := range routerIds {
			returnRouterId, err := deleteNeutronRouterByRouterId(routerId, projectId)
			if err != nil {
				return nil, err
			}
			if returnRouterId != routerId {
				log.Printf("Router delete fail %s", routerId)
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

	return returnJson, nil
}
