package activities

import (
	"context"
	"encoding/json"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-network/entities"
	"github.com/futurewei-cloud/merak/services/merak-network/http"
	"github.com/futurewei-cloud/merak/services/merak-network/utils"
	"log"
	"sync"
)

func RegisterNode(ctx context.Context, compute []*pb.InternalComputeInfo, wg *sync.WaitGroup) (string, error) {
	log.Println("RegisterNode")
	//defer wg.Done()
	log.Printf("compute %s", compute)
	nodeInfo := entities.NodeStruct{}

	for _, host := range compute {
		log.Printf("host %s", host)
		nodeBody := entities.NodeBody{
			LocalIP:    host.Ip,
			MacAddress: host.Mac,
			NodeID:     host.Id,
			NodeName:   host.Name,
			ServerPort: 50001,
			Veth:       host.Veth,
		}
		nodeInfo.Hosts = append(nodeInfo.Hosts, nodeBody)
	}
	log.Printf("nodeInfo: %s", nodeInfo)
	returnMessage, returnErr := http.RequestCall("http://"+utils.ALCORURL+":30007/nodes/bulk", "POST", nodeInfo)
	if returnErr != nil {
		log.Fatalf("returnErr %s", returnErr)
	}
	log.Printf("returnMessage %s", returnMessage)
	var returnJson entities.NodeReturn
	json.Unmarshal([]byte(returnMessage), &returnJson)
	log.Printf("returnJson : %+v", returnJson)
	log.Println("doVPC done")
	defer wg.Done()
	return "", nil
}
