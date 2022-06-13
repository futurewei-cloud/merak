package activities

import (
	"container/list"
	"context"
	"fmt"
	"github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/google/uuid"
	"log"
)

func VnetCreate(ctx context.Context, network *merak.InternalNetworkInfo) (string, error) {
	log.Printf("Test")
	log.Println("VnetCreate")
	log.Printf("merak.InternalNetworkInfo: %s", network)

	vpcs := list.New()
	for i := 0; i < int(network.NumberOfVpcs); i++ {
		id := uuid.New()
		log.Printf("UUID: %s", id.String())
		vpcs.PushBack(id.String())
	}
	for e := vpcs.Front(); e != nil; e = e.Next() {
		fmt.Println(e)
		//log.Printf(e)
	}
	return "VnetCreate", nil
}
