/*
MIT License
Copyright(c) 2022 Futurewei Cloud
    Permission is hereby granted,
    free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
    including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
    to whom the Software is furnished to do so, subject to the following conditions:
    The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
    WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/
package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
)

func GenUUID() string {
	uuidWithHyphen, _ := uuid.NewRandom()
	return strings.Replace(uuidWithHyphen.String(), "-", "", -1)
}

func ip_gen(vhost_idx int, data_plane_cidr string, upper int) string {
	var ip string = data_plane_cidr
	switch mask := strings.Split(data_plane_cidr, "/")[1]; mask {
	case "8":
		//
	case "24":
		//
	default: //16
		i := vhost_idx % upper
		ip = strings.Split(data_plane_cidr, ".")[0] + "." + strings.Split(data_plane_cidr, ".")[1] + "." + strconv.FormatInt(int64(i), 10) + "." + strconv.FormatInt(int64(vhost_idx), 10) + "/16"
	}

	return ip
}

func create_vswitches(racks []database.Vnode, ports_per_vswitch int, uid_initial int) (error, []database.Vnode, []database.Vnode) {
	var vswitches []database.Vnode
	var racks_attached []database.Vnode
	num_of_vs := len(racks) / ports_per_vswitch

	if len(racks)%ports_per_vswitch > 0 {
		num_of_vs = num_of_vs + 1
	}

	j := 0
	for i := 0; i < num_of_vs; i++ {
		err, vswitch, rs_out := create_and_attach_a_vswitch(racks, j, ports_per_vswitch, uid_initial)
		if err != nil {
			fmt.Printf("fail to create and attach vswitch %s", err)
		}

		j = j + ports_per_vswitch
		vswitches = append(vswitches, vswitch)
		racks_attached = append(racks_attached, rs_out...)
	}

	return nil, vswitches, racks_attached
}

func create_vhosts(num int, data_plane_cidr string, upper int) (error, []database.Vnode) {
	var vhosts []database.Vnode

	for j := 0; j < num; j++ {
		var vhost database.Vnode
		vhost.Id = GenUUID()
		vhost.Type = "vhost"
		vhost.Name = "vhost-" + strconv.FormatInt(int64(j), 10)
		err, nics := vhost_nics_gen(j, data_plane_cidr, upper)
		if err != nil {
			return fmt.Errorf("generate vhost nics error %s", err), nil
		} else {
			vhost.Nics = nics
		}
		vhosts = append(vhosts, vhost)

	}

	return nil, vhosts

}

func rack_nics_gen(idx int, vhosts_per_rack int) (error, []database.Nic) {
	var nics []database.Nic

	for nintf := 1; nintf <= (vhosts_per_rack + 1); nintf++ {
		var nic database.Nic
		nic.Id = GenUUID()
		nic.Intf = "r" + strconv.FormatInt(int64(idx), 10) + "-eth" + strconv.FormatInt(int64(nintf), 10)

		nics = append(nics, nic)
	}

	return nil, nics
}

func vhost_nics_gen(idx int, data_plane_cidr string, upper int) (error, []database.Nic) {
	var nics []database.Nic

	for nintf := 1; nintf <= 2; nintf++ {
		var nic database.Nic
		nic.Id = GenUUID()
		nic.Intf = "vh" + strconv.FormatInt(int64(idx), 10) + "-eth" + strconv.FormatInt(int64(nintf), 10)
		nic.Ip = ip_gen(idx, data_plane_cidr, upper)

		nics = append(nics, nic)
	}

	return nil, nics

}

func vswitch_nics_gen(idx int, ports_per_vswitch int) (error, []database.Nic) {
	var nics []database.Nic

	for nintf := 1; nintf <= (ports_per_vswitch + 1); nintf++ {
		var nic database.Nic
		nic.Id = GenUUID()
		nic.Intf = "vs" + strconv.FormatInt(int64(idx), 10) + "-eth" + strconv.FormatInt(int64(nintf), 10)

		nics = append(nics, nic)

	}

	return nil, nics

}

func core_nics_gen(idx int, nports int) (error, []database.Nic) {
	var nics []database.Nic

	for nintf := 1; nintf <= (nports); nintf++ {
		var nic database.Nic
		nic.Id = GenUUID()
		nic.Intf = "vs" + strconv.FormatInt(int64(idx), 10) + "-eth" + strconv.FormatInt(int64(nintf), 10)

		nics = append(nics, nic)

	}

	return nil, nics

}

func create_a_rack(idx int, vhosts_per_rack int) (error, database.Vnode) {

	var rack database.Vnode
	rack.Type = "rack"
	rack.Name = "rack-" + strconv.FormatInt(int64(idx), 10)
	err1, nics := rack_nics_gen(idx, vhosts_per_rack)
	if err1 != nil {
		fmt.Printf("generate rack nics error %s", err1)
	}

	rack.Nics = nics

	return nil, rack

}

func create_and_attach_a_vswitch(vs []database.Vnode, j int, ports_per_vswitch int, uid int) (error, database.Vnode, []database.Vnode) {
	var vswitch database.Vnode

	vswitch.Type = "vswitch"
	vswitch.Id = GenUUID()
	vswitch.Name = "vsu-" + strconv.FormatInt(int64(j), 10)
	err_create, nics := vswitch_nics_gen(j, ports_per_vswitch)
	if err_create != nil {
		fmt.Printf("generate uvs nics error %s", err_create)
	}
	vswitch.Nics = nics

	// attach vs to the vswitch
	err_attach, vswitch_attached, vs_attached := attach_vhosts_to_rack(vswitch, vs, uid)

	if err_attach != nil {
		fmt.Printf("attach vswitch to vs error %s", err_attach)
	}

	return nil, vswitch_attached, vs_attached
}

func create_and_attach_a_core(vs []database.Vnode, j int, nports int, uid int) (error, database.Vnode, []database.Vnode) {
	var core database.Vnode

	core.Type = "core"
	core.Id = GenUUID()
	core.Name = "core-" + strconv.FormatInt(int64(j), 10)
	err_create, nics := core_nics_gen(1, nports)
	if err_create != nil {
		fmt.Printf("generate uvs nics error %s", err_create)
	}
	core.Nics = nics

	// attach vs to the vswitch
	err_attach, core_attached, vs_attached := attach_vhosts_to_rack(core, vs, uid)

	if err_attach != nil {
		fmt.Printf("attach vswitch to vs error %s", err_attach)
	}

	return nil, core_attached, vs_attached
}

func attach_vhosts_to_rack(rack database.Vnode, hosts []database.Vnode, uid_initial int) (error, database.Vnode, []database.Vnode) {

	var rack_links []database.Vlink
	var hosts_attached []database.Vnode

	i := 0
	uid := strconv.FormatInt(int64(uid_initial), 10)

	for _, nic := range rack.Nics {

		if i < len(hosts) {
			var link_r database.Vlink
			var link_h database.Vlink

			link_r.Id = GenUUID()
			link_r.Uid = uid
			link_r.Local_pod = rack.Name
			link_r.Local_intf = nic.Intf
			link_r.Peer_pod = hosts[0].Name
			link_r.Peer_intf = hosts[0].Nics[0].Intf
			link_r.Peer_ip = hosts[0].Nics[0].Ip

			rack_links = append(rack_links, link_r)

			link_h.Id = GenUUID()
			link_h.Uid = uid
			link_h.Local_pod = hosts[i].Name
			link_h.Local_intf = hosts[i].Nics[0].Intf
			link_h.Local_ip = hosts[i].Nics[0].Ip
			link_h.Peer_pod = rack.Name
			link_h.Peer_intf = nic.Intf

			host := hosts[i]
			host.Flinks = []database.Vlink{link_h}
			hosts_attached = append(hosts_attached, host)

			i = i + 1

		}

	}

	rack.Flinks = rack_links

	return nil, rack, hosts_attached
}

func Create_multiple_layers_vswitches(vhost_num int, rack_num int, vhosts_per_rack int, ports_per_vswitch int, data_plane_cidr string) (error, database.TopologyData) {
	var topo database.TopologyData
	upper := 250
	nvhosts := vhost_num
	idx := 0
	var racks_full_attached []database.Vnode
	var racks []database.Vnode
	var vhosts []database.Vnode
	var vswitches []database.Vnode

	var core []database.Vnode

	uid_initial := 1

	for nvhosts > 0 && idx < rack_num {
		var vhs []database.Vnode
		err, rack := create_a_rack(idx, vhosts_per_rack)
		if err != nil {
			fmt.Printf("create a vswitch error %s", err)
		}
		idx = idx + 1
		if nvhosts > vhosts_per_rack {
			err1, out := create_vhosts(vhosts_per_rack, data_plane_cidr, upper)
			if err1 != nil {
				fmt.Printf("create vhosts error %s", err1)
			} else {
				vhs = append(vhs, out...)
				nvhosts = nvhosts - vhosts_per_rack

			}

		} else {
			err1, out := create_vhosts(nvhosts, data_plane_cidr, upper)
			if err1 != nil {
				fmt.Printf("create vhosts error %s", err1)
			} else {
				vhs = append(vhs, out...)
				nvhosts = nvhosts - nvhosts
			}
		}
		err, rack_host_attached, vhs_attached := attach_vhosts_to_rack(rack, vhs, uid_initial)
		if err != nil {
			fmt.Printf("attach vhosts to rack error %s", err)
		}
		uid_initial = uid_initial + len(vhs_attached)
		racks = append(racks, rack_host_attached)
		vhosts = append(vhosts, vhs_attached...)
	}

	err, vs_attached, racks_vs_attached := create_vswitches(racks, ports_per_vswitch, uid_initial)
	uid_initial = uid_initial + len(racks)
	if err != nil {
		fmt.Printf("create vswitches error %s", err)
	}

	racks_full_attached = append(racks_full_attached, racks_vs_attached...)

	nvswitch := len(vs_attached)

	var vs_to_core []database.Vnode

	for nvswitch > ports_per_vswitch {
		err_vs, vs_upper_attached, vs_lower_attached := create_vswitches(vs_attached, ports_per_vswitch, uid_initial)

		if err_vs != nil {
			fmt.Printf("create upper layer vswitches error %s", err_vs)
		}
		uid_initial = uid_initial + len(vs_attached)

		vswitches = append(vswitches, vs_lower_attached...)
		nvswitch = len(vs_upper_attached)

		if nvswitch < ports_per_vswitch {
			vs_to_core = append(vs_to_core, vs_upper_attached...)
		}

	}

	err_core, core_attached, vs_upper_attached := create_and_attach_a_core(vs_to_core, 1, len(vs_to_core), uid_initial)
	if err_core != nil {
		fmt.Printf("create and attach a core error %s", err_core)
	}
	uid_initial = uid_initial + len(vs_to_core)

	core = append(core, core_attached)
	vswitches = append(vswitches, vs_upper_attached...)

	vnodes := append(vhosts, racks_full_attached...)
	vnodes = append(vnodes, vswitches...)
	vnodes = append(vnodes, core...)

	topo.Vnodes = vnodes

	return nil, topo

}
