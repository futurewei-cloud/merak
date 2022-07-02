package tests

import (
	"fmt"
	"testing"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	"github.com/futurewei-cloud/merak/services/merak-topo/handler"
	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
)

var (
	returnMessage = pb.ReturnTopologyMessage{
		ReturnCode:    pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

func TestTopologyHandler(t *testing.T) {
	aca_num := 2
	rack_num := 1
	aca_per_rack := 2
	data_plane_cidr := "10.200.0.0/16"
	// topo_id := "topo0001"

	// redis init check-- pingpong test

	k8client, err := utils.K8sClient()
	if err != nil {
		fmt.Printf("create k8s client error %s", err)
	}

	err1 := database.ConnectDatabase()
	if err1 != nil {
		fmt.Printf("connect to DB error %s", err1)
	}

	err2 := handler.Create(k8client, uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), data_plane_cidr, &returnMessage)
	if err2 != nil {
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Fail to Create Topology."

	} else {
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Topology Deployed."

	}

	fmt.Printf("///// Return Message //// %v", &returnMessage)

	// err = handler.Delete(k8client,topo_id)
	// if err !=nil {
	// 	fmt.Errorf("delete topology error %s", err)
	// }

	// handler.Subtest(k8client, topo_id)

}
