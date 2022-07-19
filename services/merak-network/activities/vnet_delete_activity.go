package activities

import (
	"context"
	"encoding/json"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-network/database"
	"github.com/futurewei-cloud/merak/services/merak-network/http"
	"github.com/futurewei-cloud/merak/services/merak-network/utils"
	"log"
	"sync"
)

func deleteSubnet(subnetId string) (returnSubnetId string) {
	log.Println("deleteSubnet")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30002/project/123456789/subnets/"+subnetId, "DELETE", "")
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	//database.Del(utils.SUBNET + subnetId)
	log.Println("deleteSubnet done")
	return subnetId
}

func deleteVpc(vpcId string) (returnVpcId string) {
	log.Println("deleteVpc")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30001/project/123456789/vpcs/"+vpcId, "DELETE", "")
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	//database.Del(utils.VPC + vpcId)
	log.Println("deleteSubnet done")
	return vpcId
}

func deleteRouter(vpcId string) (returnVpcId string) {
	log.Println("deleteRouter")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30003/project/123456789/vpcs/"+vpcId+"/router", "DELETE", "")
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	//database.Del(utils.Router + vpcId)
	log.Println("deleteSubnet done")
	return vpcId
}

func deleteRouteTable(subnetId string) (returnVpcId string) {
	log.Println("deleteRouteTable")
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30003/project/123456789/subnets/"+subnetId+"/routetable", "DELETE", "")
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	//database.Del(utils.Router + vpcId)
	log.Println("deleteSubnet done")
	return subnetId
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
		for _, subnet := range vpc.Subnets {
			deleteRouteTable(subnet.SubnetId)
			deleteSubnet(subnet.SubnetId)
		}
		deleteRouter(vpc.VpcId)
		deleteVpc(vpc.VpcId)
	}
	//database.Del(utils.NETCONFIG + netConfigId)

	json.Unmarshal([]byte(values), &returnJson)
	log.Printf("returnMessage %s", returnJson)

	returnMessage <- returnJson
	return "VnetInfo", nil
}
