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
		i := (vhost_idx) / upper

		ip = strings.Split(data_plane_cidr, ".")[0] + "." + strings.Split(data_plane_cidr, ".")[1] + "." + strconv.FormatInt(int64(i), 10) + "." + strconv.FormatInt(int64((vhost_idx-i*upper)), 10) + "/16"
	}

	return ip
}

func create_vswitches(racks []database.Vnode, init_idx_vs int, ports_per_vswitch int, uid_initial int) (error, []database.Vnode, []database.Vnode) {
	var vswitches []database.Vnode
	var racks_attached []database.Vnode
	num_of_vs := len(racks) / ports_per_vswitch

	if len(racks)%ports_per_vswitch > 0 {
		num_of_vs = num_of_vs + 1
	}

	j := init_idx_vs
	k := ports_per_vswitch
	for i := 0; i < num_of_vs; i++ {
		lower_bound := i * k
		var upper_bound int
		if i*k+k-1 < len(racks) {
			upper_bound = lower_bound + k
		} else {
			upper_bound = len(racks)
		}

		err, vswitch, rs_out := create_and_attach_a_vswitch(racks[lower_bound:upper_bound], j, ports_per_vswitch, uid_initial)
		if err != nil {
			return fmt.Errorf("fail to create and attach vswitch %s", err), nil, nil
		}

		vswitches = append(vswitches, vswitch)
		racks_attached = append(racks_attached, rs_out...)
		uid_initial = uid_initial + ports_per_vswitch
		j = j + 1
	}

	return nil, vswitches, racks_attached
}

func create_vhosts(init int, vhosts_per_rack int, data_plane_cidr string, upper int) []database.Vnode {
	var vhosts []database.Vnode

	for j := init; j < init+vhosts_per_rack; j++ {
		var vhost database.Vnode
		vhost.Id = GenUUID()
		vhost.Type = "vhost"
		vhost.Name = "vhost-" + strconv.FormatInt(int64(j), 10)
		nics := vhost_nics_gen(j, data_plane_cidr, upper)
		vhost.Nics = nics

		vhosts = append(vhosts, vhost)

	}

	return vhosts

}

func rack_nics_gen(idx int, vhosts_per_rack int) []database.Nic {
	var nics []database.Nic

	for nintf := 1; nintf <= (vhosts_per_rack + 1); nintf++ {
		var nic database.Nic
		nic.Id = GenUUID()
		nic.Intf = "r" + strconv.FormatInt(int64(idx), 10) + "-eth" + strconv.FormatInt(int64(nintf), 10)

		nics = append(nics, nic)
	}

	return nics
}

func vhost_nics_gen(idx int, data_plane_cidr string, upper int) []database.Nic {
	var nics []database.Nic

	var nic database.Nic
	nic.Id = GenUUID()
	nic.Intf = "vh" + strconv.FormatInt(int64(idx), 10) + "-eth1"
	nic.Ip = ip_gen(idx, data_plane_cidr, upper)

	nics = append(nics, nic)

	return nics

}

func vswitch_nics_gen(idx int, ports_per_vswitch int) []database.Nic {
	var nics []database.Nic

	for nintf := 1; nintf <= (ports_per_vswitch + 1); nintf++ {
		var nic database.Nic
		nic.Id = GenUUID()
		nic.Intf = "vs" + strconv.FormatInt(int64(idx), 10) + "-eth" + strconv.FormatInt(int64(nintf), 10)

		nics = append(nics, nic)

	}

	return nics

}

func core_nics_gen(idx int, nports int) []database.Nic {
	var nics []database.Nic

	for nintf := 1; nintf < (nports + 1); nintf++ {
		var nic database.Nic
		nic.Id = GenUUID()
		nic.Intf = "c" + strconv.FormatInt(int64(idx), 10) + "-eth" + strconv.FormatInt(int64(nintf), 10)

		nics = append(nics, nic)

	}

	return nics

}

func create_a_rack(idx int, vhosts_per_rack int) database.Vnode {

	var rack database.Vnode
	rack.Type = "rack"
	rack.Name = "rack-" + strconv.FormatInt(int64(idx), 10)
	nics := rack_nics_gen(idx, vhosts_per_rack)

	rack.Nics = nics

	return rack

}

func create_and_attach_a_vswitch(vs []database.Vnode, idx_vs int, ports_per_vswitch int, uid_initial int) (error, database.Vnode, []database.Vnode) {
	var vswitch database.Vnode
	var nports int

	vswitch.Type = "vswitch"
	vswitch.Id = GenUUID()
	vswitch.Name = "vs-" + strconv.FormatInt(int64(idx_vs), 10)

	if ports_per_vswitch <= len(vs) {
		nports = ports_per_vswitch
	} else {
		nports = len(vs)
	}

	nics := vswitch_nics_gen(idx_vs, nports)

	vswitch.Nics = nics

	err_attach, vswitch_attached, vs_attached := attach_racks_to_vswitch(vswitch, vs, uid_initial)

	if err_attach != nil {
		fmt.Printf("attach vswitch to vs error %s", err_attach)
	}

	return err_attach, vswitch_attached, vs_attached
}

func create_and_attach_a_core(vs []database.Vnode, j int, nports int, uid_initial int) (error, database.Vnode, []database.Vnode) {
	var core database.Vnode

	core.Type = "core"
	core.Id = GenUUID()
	core.Name = "core-" + strconv.FormatInt(int64(j), 10)
	nics := core_nics_gen(1, nports)

	core.Nics = nics

	err_attach, core_attached, vs_attached := attach_vswitches_to_core(core, vs, uid_initial)

	if err_attach != nil {
		fmt.Printf("attach vswitch to vs error %s", err_attach)
	}

	return err_attach, core_attached, vs_attached
}

func attach_vswitches_to_core(core database.Vnode, vswitches []database.Vnode, uid_initial int) (error, database.Vnode, []database.Vnode) {

	var core_links []database.Vlink
	var vswitches_attached []database.Vnode

	for i, nic := range core.Nics {
		var link_c database.Vlink
		var link_v database.Vlink

		uid := uid_initial + i

		link_c.Id = GenUUID()
		link_c.Uid = uid
		link_c.Name = core.Name + "-l" + strconv.FormatInt(int64(uid), 10)
		link_c.Local_pod = core.Name
		link_c.Local_intf = nic.Intf
		link_c.Peer_pod = vswitches[i].Name
		link_c.Peer_intf = vswitches[i].Nics[len(vswitches[i].Nics)-1].Intf

		core_links = append(core_links, link_c)

		link_v.Id = GenUUID()
		link_v.Uid = uid
		link_v.Name = vswitches[i].Name + "-l" + strconv.FormatInt(int64(uid), 10)
		link_v.Local_pod = vswitches[i].Name
		link_v.Local_intf = vswitches[i].Nics[len(vswitches[i].Nics)-1].Intf
		link_v.Peer_pod = core.Name
		link_v.Peer_intf = nic.Intf

		vswitch := vswitches[i]
		vswitch.Flinks = append(vswitch.Flinks, link_v)
		vswitches_attached = append(vswitches_attached, vswitch)

	}

	core.Flinks = core_links

	return nil, core, vswitches_attached
}

func attach_vhosts_to_rack(rack database.Vnode, hosts []database.Vnode, uid_initial int) (error, database.Vnode, []database.Vnode) {

	var rack_links []database.Vlink
	var hosts_attached []database.Vnode

	for i, nic := range rack.Nics {

		if i < len(rack.Nics)-1 {
			var link_r database.Vlink
			var link_h database.Vlink

			uid := uid_initial + i

			link_r.Id = GenUUID()
			link_r.Uid = uid
			link_r.Name = rack.Name + "-l" + strconv.FormatInt(int64(uid), 10)
			link_r.Local_pod = rack.Name
			link_r.Local_intf = nic.Intf
			link_r.Peer_pod = hosts[i].Name
			link_r.Peer_intf = hosts[i].Nics[0].Intf
			link_r.Peer_ip = hosts[i].Nics[0].Ip

			rack_links = append(rack_links, link_r)

			link_h.Id = GenUUID()
			link_h.Uid = uid
			link_h.Name = hosts[i].Name + "-l" + strconv.FormatInt(int64(uid), 10)
			link_h.Local_pod = hosts[i].Name
			link_h.Local_intf = hosts[i].Nics[0].Intf
			link_h.Local_ip = hosts[i].Nics[0].Ip
			link_h.Peer_pod = rack.Name
			link_h.Peer_intf = nic.Intf

			host := hosts[i]
			host.Flinks = append(host.Flinks, link_h)
			hosts_attached = append(hosts_attached, host)

		}

	}

	rack.Flinks = rack_links
	return nil, rack, hosts_attached
}

func attach_racks_to_vswitch(vswitch database.Vnode, racks []database.Vnode, uid_initial int) (error, database.Vnode, []database.Vnode) {

	var vswitch_links []database.Vlink
	var racks_attached []database.Vnode

	for i, nic := range vswitch.Nics {
		if i < len(vswitch.Nics)-1 {
			var link_v database.Vlink
			var link_r database.Vlink

			uid := uid_initial + i

			link_v.Id = GenUUID()
			link_v.Uid = uid
			link_v.Name = vswitch.Name + "-l" + strconv.FormatInt(int64(uid), 10)
			link_v.Local_pod = vswitch.Name
			link_v.Local_intf = nic.Intf
			link_v.Peer_pod = racks[i].Name
			link_v.Peer_intf = racks[i].Nics[len(racks[i].Nics)-1].Intf

			vswitch_links = append(vswitch_links, link_v)

			link_r.Id = GenUUID()
			link_r.Uid = uid
			link_r.Name = racks[i].Name + "-l" + strconv.FormatInt(int64(uid), 10)
			link_r.Local_pod = racks[i].Name
			link_r.Local_intf = racks[i].Nics[len(racks[i].Nics)-1].Intf
			link_r.Peer_pod = vswitch.Name
			link_r.Peer_intf = nic.Intf

			rack := racks[i]
			rack.Flinks = append(rack.Flinks, link_r)
			racks_attached = append(racks_attached, rack)

		}

	}

	vswitch.Flinks = vswitch_links

	return nil, vswitch, racks_attached
}

func Create_multiple_layers_vswitches(vhost_num int, rack_num int, vhosts_per_rack int, ports_per_vswitch int, data_plane_cidr string) (error, database.TopologyData) {
	var topo database.TopologyData
	upper := 250
	nvhosts := vhost_num
	idx := 1
	var racks_full_attached []database.Vnode
	var racks []database.Vnode
	var vhosts []database.Vnode
	var vswitches []database.Vnode

	uid_initial := 1
	init_idx_host := 1
	init_idx_vs := 1

	for nvhosts > 0 && idx < rack_num+1 {
		var vhs []database.Vnode
		rack := create_a_rack(idx, vhosts_per_rack)

		idx = idx + 1

		if nvhosts > vhosts_per_rack {
			out := create_vhosts(init_idx_host, vhosts_per_rack, data_plane_cidr, upper)
			vhs = append(vhs, out...)
			nvhosts = nvhosts - vhosts_per_rack
			init_idx_host = init_idx_host + vhosts_per_rack

		} else {
			out := create_vhosts(init_idx_host, nvhosts, data_plane_cidr, upper)
			vhs = append(vhs, out...)
			nvhosts = 0
			init_idx_host = init_idx_host + vhosts_per_rack

		}
		err, rack_host_attached, vhs_attached := attach_vhosts_to_rack(rack, vhs, uid_initial)
		if err != nil {
			fmt.Printf("attach vhosts to rack error %s", err)
		}
		uid_initial = uid_initial + len(vhs_attached)
		racks = append(racks, rack_host_attached)
		vhosts = append(vhosts, vhs_attached...)
	}

	err, vs_attached, racks_vs_attached := create_vswitches(racks, init_idx_vs, ports_per_vswitch, uid_initial)
	uid_initial = uid_initial + len(racks)
	init_idx_vs = init_idx_vs + len(vs_attached)
	if err != nil {
		fmt.Printf("create vswitches error %s", err)
	}

	racks_full_attached = append(racks_full_attached, racks_vs_attached...)

	nvswitch := len(vs_attached)

	flag := false

	var vs_to_core []database.Vnode

	for nvswitch > ports_per_vswitch {
		flag = true
		err_vs, vs_upper_attached, vs_lower_attached := create_vswitches(vs_attached, init_idx_vs, ports_per_vswitch, uid_initial)

		if err_vs != nil {
			fmt.Printf("create upper layer vswitches error %s", err_vs)
		}
		uid_initial = uid_initial + len(vs_attached)
		init_idx_vs = init_idx_vs + len(vs_upper_attached)

		vswitches = append(vswitches, vs_lower_attached...)
		nvswitch = len(vs_upper_attached)

		vs_to_core = append(vs_to_core, vs_upper_attached...)

	}

	if !flag {
		vs_to_core = vs_attached
	}

	err_core, core_attached, vs_attached := create_and_attach_a_core(vs_to_core, 1, len(vs_to_core), uid_initial)
	if err_core != nil {
		fmt.Printf("create and attach a core error %s", err_core)
	}

	vswitches = append(vswitches, vs_attached...)

	vnodes := append(vhosts, racks_full_attached...)
	vnodes = append(vnodes, vswitches...)
	vnodes = append(vnodes, core_attached)

	topo.Vnodes = vnodes

	return nil, topo

}
