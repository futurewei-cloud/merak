package tests

// import (
// 	"fmt"
// 	"testing"

// 	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
// 	"github.com/futurewei-cloud/merak/services/merak-topo/database"
// 	"github.com/futurewei-cloud/merak/services/merak-topo/handler"
// 	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
// )

// var (
// 	returnMessage = pb.ReturnTopologyMessage{
// 		ReturnCode:    pb.ReturnCode_FAILED,
// 		ReturnMessage: "Unintialized",
// 	}
// 	aca_num         = 100
// 	rack_num        = 10
// 	aca_per_rack    = 10
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
