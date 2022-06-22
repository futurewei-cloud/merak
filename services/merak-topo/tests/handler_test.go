package tests

import (
	"fmt"
	"testing"

	"github.com/futurewei-cloud/merak/services/merak-topo/handler"

	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
)

func TestTopologyHandler(t *testing.T) {
	aca_num := 4
	rack_num := 2
	aca_per_rack := 2
	data_plane_cidr := "10.200.0.0/16"

	k8client, err := utils.K8sClient()
	if err != nil {
		fmt.Printf("create k8s client error %s", err)
	}

	topology_create := handler.Create(k8client, uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), data_plane_cidr)

	fmt.Printf("The created topology is: %+v \n", topology_create)
}
