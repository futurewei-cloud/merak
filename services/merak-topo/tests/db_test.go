package tests

import (
	"fmt"

	"testing"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
)

func TestDB(t *testing.T) {
	t_client := database.CreateDBClient()
	fmt.Println("DB created.")

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

	vnode0 := database.Vnode{
		Name: "a1",
		Nics: []database.Nic{nic1, nic2},
	}

	vnode1 := database.Vnode{
		Name: "a2",
		Nics: []database.Nic{nic3},
	}

	local1 := database.Vport{
		Name: vnode1.Name,
		Intf: vnode1.Nics[len(vnode1.Nics)-1].Intf,
		Ip:   vnode1.Nics[len(vnode1.Nics)-1].Ip,
	}
	peer1 := database.Vport{
		Name: vnode0.Name,
		Intf: vnode0.Nics[len(vnode0.Nics)-1].Intf,
		Ip:   vnode0.Nics[len(vnode0.Nics)-1].Ip,
	}

	vlink1 := database.Vlink{
		Name: "link1",
		Src:  local1,
		Dst:  peer1,
	}

	// vlink2 := database.Vlink{
	// 	Name:  "link2",
	// 	Local: local2,
	// 	Peer:  peer2,
	// }

	topo1 := database.TopologyData{
		Topology_id: "topo1",
		Vnodes:      []database.Vnode{vnode0, vnode1},
		Vlinks:      []database.Vlink{vlink1},
	}

	fmt.Println("set topo1")
	database.SetValue(t_client, "topo:proj1-topo1", topo1)
	fmt.Println("set node1")
	database.SetValue(t_client, "node:proj1-topo1-node1", vnode0)
	fmt.Println("set link1")
	database.SetValue(t_client, "link:proj1-topo1-link1", vlink1)

	// vnode--- key   vnode:proj1-topo1-vnode1
	// vlink --- key  vlink:proj1-topo1-vlink1

	val := database.GetValue(t_client, "topo:proj1-topo1")

	fmt.Printf("The returned data from Radis is %+v. \n", val)

	val = database.GetValue(t_client, "node:proj1-topo1-node1")

	fmt.Printf("The returned data from Radis is %+v.\n ", val)

	val = database.GetValue(t_client, "link:proj1-topo1-link1")

	fmt.Printf("The returned data from Radis is %+v.\n", val)

}
