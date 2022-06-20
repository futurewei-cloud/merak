package handler

import (
	"fmt"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
)

// grpc client

// Call operation.Create()
// --- Generete topology yaml file based on topology type

// Call operation.SaveTopo()
// ----Save topology yaml file

// Call operation.ConfigTopo()
// ---generate pod configuration data for creating pods
// ---generate config-map for each pod in the topology through volum_amount
// ---generate pod specification

// Call operation. DeployTopo()
// ---through k8s pod creation
// ---through k8s deployment

var (
	Topo    database.TopologyData
	Vlinks  []database.Vlink
	Vnodes  []database.Vnode
	cgw_num int = 2
	count   int = 250
	k       int = 0 // subnet starting number
)

//function CREATE
func Create(aca_num uint32, rack_num uint32, aca_per_rack uint32, data_plane_cidr string) database.TopologyData {

	// topo-gen
	var ovs_tor_device = []string{"tor-0"}
	ip_num := int(aca_num) + cgw_num

	ips := Ips_gen(ip_num, k, count, data_plane_cidr)

	fmt.Println("=== parse done == ")
	fmt.Printf("TOR OVS is: %v \n", ovs_tor_device[0])
	fmt.Printf("Vswitch number is: %v\n", rack_num)
	fmt.Printf("Vhost number is: %v\n", aca_num)

	fmt.Println("======== Generate device list ==== ")
	rack_device := Pod_name(int(rack_num), "vswitch")
	aca_device := Pod_name(int(aca_num), "vhost")
	ngw_device := Pod_name(cgw_num, "cgw")

	fmt.Printf("Vswitch_device: %v\n", rack_device)
	fmt.Printf("Vhost_device: %v\n", aca_device)
	fmt.Printf("Cgw_device: %v\n", ngw_device)

	fmt.Println("======== Generate device nodes ==== ")
	rack_intf_num := int(aca_per_rack) + 1
	tor_intf_num := int(rack_num) + cgw_num
	aca_intf_num := 1
	ngw_intf_num := 1

	Node_port_gen(rack_intf_num, rack_device, "vswitch", ips, false)
	ips_1 := Node_port_gen(aca_intf_num, aca_device, "vhost", ips, true)
	Node_port_gen(tor_intf_num, ovs_tor_device, "tor", ips, false)
	Node_port_gen(ngw_intf_num, ngw_device, "cgw", ips_1, true)

	fmt.Printf("The topology nodes are : %+v. \n", Topo_nodes)

	// cur_picked_intfs := []string{}

	fmt.Println("======== Pairing links ==== ")
	// left_ports_tor, new_picked, new_picked_intfs := Link_gen(ngw_ports, tor_ports, cur_picked, cur_picked_intfs)
	// left_ports_rack, new_picked1, new_picked_intfs1 := Link_gen(left_ports_tor, rack_ports, new_picked, new_picked_intfs)
	// Link_gen(left_ports_rack, aca_ports, new_picked1, new_picked_intfs1)
	Topo.Topology_id = "topo:" + GenUUID()

	Links_gen(Topo_nodes, Topo.Topology_id)
	fmt.Printf("The topology links are : %+v. \n", Topo_links)

	fmt.Println("======== Generate topology data ==== ")

	Topo.Vlinks = Topo_links
	Topo.Vnodes = Topo_nodes

	// fmt.Println("======== Topology Deployment ==== ")
	// Topo_deploy()

	return Topo

	// topo-deploy

	// save to radis

	// network config
	// --rack ovs config
	// --routing config
	// --test all ping

}

func Delete() {

}

func Update() {

}
