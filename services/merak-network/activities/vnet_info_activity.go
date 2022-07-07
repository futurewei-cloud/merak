package activities

import (
	"context"
	"encoding/json"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-network/database"
	"github.com/futurewei-cloud/merak/services/merak-network/utils"
	"log"
	"sync"
)

func VnetInfo(ctx context.Context, netConfigId string, wg *sync.WaitGroup, returnMessage chan *pb.ReturnNetworkMessage) (string, error) {
	defer wg.Done()
	// TODO: when query db, make sure to check if key exist first, other wise could timeout
	log.Println("VnetInfo")
	values, err := database.Get(utils.NETCONFIG + netConfigId)
	if err != nil {
		return "", err
	}
	log.Printf("VnetInfo %s", values)
	log.Printf("returnMessage %s", returnMessage)
	var returnJson *pb.ReturnNetworkMessage
	json.Unmarshal([]byte(values), &returnJson)
	log.Printf("returnMessage %s", returnJson)
	returnMessage <- returnJson
	return "VnetInfo", nil
}
