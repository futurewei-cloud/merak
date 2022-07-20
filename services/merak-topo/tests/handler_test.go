package tests

import (
	"fmt"
	"testing"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	"github.com/futurewei-cloud/merak/services/merak-topo/handler"
	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
)

// var (
// 	returnMessage = pb.ReturnTopologyMessage{
// 		ReturnCode:    pb.ReturnCode_FAILED,
// 		ReturnMessage: "Unintialized",
// 	}
// 	aca_num         = 15
// 	rack_num        = 3
// 	aca_per_rack    = 5
// 	data_plane_cidr = "10.200.10.0/16"
// 	topo_id         = "20"
// 	cgw_num         = 2
// )

// func TestTopologyCreate(t *testing.T) {

// 	k8client, err := utils.K8sClient()
// 	if err != nil {
// 		fmt.Printf("create k8s client error %s", err)
// 	}

// 	err1 := database.ConnectDatabase()
// 	if err1 != nil {
// 		fmt.Printf("connect to DB error %s", err1)
// 	}

// 	err2 := handler.Create(k8client, topo_id, uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), uint32(cgw_num), data_plane_cidr, &returnMessage)
// 	if err2 != nil {
// 		returnMessage.ReturnCode = pb.ReturnCode_FAILED
// 		returnMessage.ReturnMessage = "Fail to Create Topology."

// 	} else {
// 		returnMessage.ReturnCode = pb.ReturnCode_OK
// 		returnMessage.ReturnMessage = "Topology Deployed."

// 	}

// 	fmt.Printf("///// CREATE Return Message //// %v", &returnMessage)

// }

func TestTopologyInfo(t *testing.T) {

	// topo_id := "1234560002"
	// aca_num := 10

	k8client, err := utils.K8sClient()
	if err != nil {
		fmt.Printf("create k8s client error %s", err)
	}

	err1 := database.ConnectDatabase()
	if err1 != nil {
		fmt.Printf("connect to DB error %s", err1)
	}

	err3 := handler.Info(k8client, topo_id, (aca_num + cgw_num), &returnMessage)

	if err3 != nil {
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Topology Info Fails"

	} else {
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Topology Info Passes."
	}

	fmt.Printf("///// INFO Return Message //// %v", &returnMessage)

}

func TestTopologyDelete(t *testing.T) {

	k8client, err := utils.K8sClient()
	if err != nil {
		fmt.Printf("create k8s client error %s", err)
	}

	err1 := database.ConnectDatabase()
	if err1 != nil {
		fmt.Printf("connect to DB error %s", err1)
	}

	err4 := handler.Delete(k8client, topo_id)
	if err4 != nil {
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Fail to Delete Topology."
	} else {
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Topology Delete Done."
	}

	fmt.Printf("///// DELETE Return Message //// %v", &returnMessage)

}

// func TestQueryMac(t *testing.T) {

// 	topo_id := "1234560001"

// 	// redis init check-- pingpong test

// 	k8client, err := utils.K8sClient()
// 	if err != nil {
// 		fmt.Printf("create k8s client error %s", err)
// 	}

// 	err1 := database.ConnectDatabase()
// 	if err1 != nil {
// 		fmt.Printf("connect to DB error %s", err1)
// 	}

// 	err2 := handler.QueryMac(k8client, topo_id)

// 	if err != nil {
// 		fmt.Printf("err on running k8s command %s", err2)
// 	}

// }

// func TestQueryHostNode(t *testing.T) {

// 	topo_id := "1234560001"

// 	k8client, err := utils.K8sClient()
// 	if err != nil {
// 		fmt.Printf("create k8s client error %s", err)
// 	}

// 	err1 := database.ConnectDatabase()
// 	if err1 != nil {
// 		fmt.Printf("connect to DB error %s", err1)
// 	}

// 	err2 := handler.QueryHostNode(k8client, topo_id)

// 	if err != nil {
// 		fmt.Printf("err on running k8s command %s", err2)
// 	}

// }

func TestTopologyHandler(t *testing.T) {

	k8client, err := utils.K8sClient()
	if err != nil {
		fmt.Printf("create k8s client error %s", err)
	}

	err1 := database.ConnectDatabase()
	if err1 != nil {
		fmt.Printf("connect to DB error %s", err1)
	}

	err2 := handler.Create(k8client, topo_id, uint32(aca_num), uint32(rack_num), uint32(aca_per_rack), uint32(cgw_num), data_plane_cidr, &returnMessage)
	if err2 != nil {
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Fail to Create Topology."

	} else {
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Topology Deployed."

	}

	fmt.Printf("///// CREATE Return Message //// %v", &returnMessage)

	err3 := handler.Info(k8client, topo_id, (aca_num + cgw_num), &returnMessage)

	if err3 != nil {
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Fail to Info Topology."

	} else {
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Topology Info Query Done."
	}

	fmt.Printf("///// INFO Return Message //// %v", &returnMessage)

	err4 := handler.Delete(k8client, topo_id)
	if err4 != nil {
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "Fail to Delete Topology."
	} else {
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "Topology Delete Done."
	}

	fmt.Printf("///// DELETE Return Message //// %v", &returnMessage)

}
