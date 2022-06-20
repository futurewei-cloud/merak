package tests

import (
	"fmt"
	"testing"

	"github.com/futurewei-cloud/merak/services/merak-topo/handler"
)

func TestTopologyHandler(t *testing.T) {
	aca_num := 16
	rack_num := 4
	aca_per_rack := 4
	data_plane_cidr := "10.200.0.0/16"

	topology_create := handler.Create(uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), data_plane_cidr)

	fmt.Printf("The created topology is: %+v \n", topology_create)
}
