package tests

import (
	"fmt"

	"testing"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
)

func TestDB(t *testing.T) {
	database.CreateDBClient()
	fmt.Println("DB created")

	//     "node":
	//         {
	//             "name": "a1"
	//             "nics": [
	//                 {
	//                     "intf": "a1-intf1"
	//                     "ip":"10.99.1.2"
	//                 },
	//                 {
	//                     "intf": "a1-intf2"
	//                     "ip":"10.99.1.3"
	//                 }
	//             ]

	//         }
	// }

	nic1 := database.Nic{
		Intf: "a1-intf1",
		Ip:   "10.99.1.2",
	}
	nic2 := database.Nic{
		Intf: "a1-intf2",
		Ip:   "10.99.1.3",
	}

	nic3 := database.Nic{
		Intf: "a2-intf1",
		Ip:   "10.99.1.4",
	}
	// nic4 := database.Nic{
	// 	Intf: "a3-intf1",
	// 	Ip:   "10.99.1.5",
	// }

	vnode0 := database.Vnode{
		Name: "a1",
		Nics: []database.Nic{nic1, nic2},
	}

	vnode1 := database.Vnode{
		Name: "a2",
		Nics: []database.Nic{nic3},
	}

	// vnode2 := database.Vnode{
	// 	Name: "a3",
	// 	Nics: []database.Nic{nic4},
	// }

	// {
	//     "link":
	//        {
	//             "name": "link1",
	//             "pairs":
	//                 {
	//                     "local_name": "a2"
	//                     "local_nics": "a2-intf1"
	//                     "local_ip": ""
	//                     "peer_name": "a3"
	//                     "peer_nics": "a3-intf1"
	//                     "peer_ip": ""
	//                 }

	//         }

	// }

	local1 := database.Vport{
		Name: vnode1.Name,
		Nics: vnode1.Nics[len(vnode1.Nics)-1].Intf,
		Ip:   vnode1.Nics[len(vnode1.Nics)-1].Ip,
	}
	peer1 := database.Vport{
		Name: vnode0.Name,
		Nics: vnode0.Nics[len(vnode0.Nics)-1].Intf,
		Ip:   vnode0.Nics[len(vnode0.Nics)-1].Ip,
	}

	// local2 := database.Vport{
	// 	Name: vnode1.Name,
	// 	Nics: vnode1.Nics[len(vnode1.Nics)-2].Intf,
	// 	Ip:   vnode1.Nics[len(vnode1.Nics)-2].Ip,
	// }

	// peer2 := database.Vport{
	// 	Name: vnode2.Name,
	// 	Nics: vnode2.Nics[len(vnode2.Nics)-1].Intf,
	// 	Ip:   vnode2.Nics[len(vnode2.Nics)-1].Ip,
	// }

	vlink1 := database.Vlink{
		Name:  "link1",
		Local: local1,
		Peer:  peer1,
	}

	// vlink2 := database.Vlink{
	// 	Name:  "link2",
	// 	Local: local2,
	// 	Peer:  peer2,
	// }

	topo1 := database.TopologyData{
		Topology_id: "topo1",
		Vnodes:      []string{"vnode0", "vnode1", "vnode2"},
		Vlinks:      []string{"vlink1", "vlink2"},
	}

	var val string

	database.SetValue("topo:proj1-topo1", topo1)
	database.SetValue("node:proj1-topo1-node1", vnode0)
	database.SetValue("link:proj1-topo1-link1", vlink1)

	// vnode--- key   vnode:proj1-topo1-vnode1
	// vlink --- key  vlink:proj1-topo1-vlink1

	val = database.GetValue("topo:proj1-topo1")

	fmt.Printf("The returned data from Radis is %+v ", val)

	val = database.GetValue("node:proj1-topo1-node1")

	fmt.Printf("The returned data from Radis is %+v ", val)

	val = database.GetValue("link:proj1-topo1-link1")

	fmt.Printf("The returned data from Radis is %+v ", val)

}
