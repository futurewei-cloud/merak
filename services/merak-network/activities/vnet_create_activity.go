package activities

import (
	"context"
	"github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/google/uuid"
	"log"
)

type router struct {
	router_id  string
	subnet_ids []string
}

type subnet struct {
	subnet_id    string
	subnet_cider string
	vpc_id       string
}

type vpc struct {
	vpc_id    string
	vpc_cider string
}

func VnetCreate(ctx context.Context, network *merak.InternalNetworkInfo) (string, error) {
	log.Printf("Test")
	log.Println("VnetCreate")
	log.Printf("merak.InternalNetworkInfo: %s", network)

	// TODO may want to separate bellow sections to different function, and use `go` and `wg` to improve overall speed
	// VPC
	vpcs := []vpc{}
	for i := 0; i < int(network.NumberOfVpcs); i++ {
		vpc_id := uuid.New().String()
		log.Printf("vpc UUID: %s", vpc_id)
		current_vpc := vpc{
			vpc_id:    vpc_id,
			vpc_cider: "hahaha",
		}
		vpcs = append(vpcs, current_vpc)
	}
	log.Printf("vpcs : %s", vpcs)

	// Subnet
	subnets := []subnet{}
	subnetCount := 0
	for i := 0; i < int(network.NumberOfVpcs); i++ {
		vpc_id := vpcs[i].vpc_id
		for j := 0; j < int(network.NumberOfSubnetPerVpc); j++ {
			subnet_id := uuid.New().String()
			log.Printf("subnet UUID: %s", subnet_id)
			current_subnet := subnet{
				subnet_id:    subnet_id,
				subnet_cider: network.SubnetCiders[subnetCount],
				vpc_id:       vpc_id,
			}
			subnets = append(subnets, current_subnet)
			subnetCount++
		}
	}

	// Router
	for _, r := range network.Routers {
		router_id := uuid.New().String()
		for _, subnet := range r.Subnets {
			current_router := router{
				router_id:  router_id,
				subnet_ids: subnet,
			}
		}
	}
	return "VnetCreate", nil
}
