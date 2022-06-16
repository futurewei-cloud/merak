package handler

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
)

var (
	Topo_links []database.Vlink
	Topo_nodes []database.Vnode
)

//function methods

func GenUUID() string {
	uuidWithHyphen, _ := uuid.NewRandom()
	return strings.Replace(uuidWithHyphen.String(), "-", "", -1)
}

func Pod_name(num int, prefix string) []string {
	var pod_name []string
	var pod_n string

	if prefix == "vswitch" {
		for i := 0; i < num; i++ {
			pod_n = prefix + "-" + strconv.FormatInt(int64(i+1), 10)
			pod_name = append(pod_name, pod_n)
		}
	} else {
		for i := 0; i < num; i++ {
			pod_n = prefix + "-" + strconv.FormatInt(int64(i), 10)
			pod_name = append(pod_name, pod_n)
		}
	}
	return pod_name
}

func Intf_name(dev_num int, prefix string) []string {
	var intf_name = []string{}
	var intf_n string = ""
	if strings.Contains(prefix, "vhost") {
		intf_n = strings.Split(prefix, "-")[0] + strings.Split(prefix, "-")[1] + "-eth1"
		intf_name = append(intf_name, intf_n)
	} else {
		for i := 0; i < dev_num; i++ {
			intf_n = strings.Split(prefix, "-")[0] + strings.Split(prefix, "-")[1] + "-eth" + strconv.FormatInt(int64(i+1), 10)
			intf_name = append(intf_name, intf_n)
		}
	}
	return intf_name
}

func ip_gen_sub(upper int, k1 int, data_plane_cidr string) []string {
	var ips_gen = []string{}
	var ip string = data_plane_cidr
	switch mask := strings.Split(data_plane_cidr, "/")[1]; mask {
	case "8":
		//
	case "24":
		//
	default: //16
		for i := 0; i < upper; i++ {
			ip = strings.Split(data_plane_cidr, ".")[0] + "." + strings.Split(data_plane_cidr, ".")[1] + "." + strconv.FormatInt(int64(k1), 10) + "." + strconv.FormatInt(int64(i+1), 10) + "/16"
			ips_gen = append(ips_gen, ip)
		}
	}

	return ips_gen
}

func Ips_gen(ip_num int, k int, count int, data_plane_cidr string) []string {
	var ips = []string{}
	left := ip_num - count
	//k-- subnet starting number

	t := int(ip_num / count)

	if t == 0 {
		ips = ip_gen_sub(ip_num, k, data_plane_cidr)
	} else if t == 1 {
		ips = ip_gen_sub(count, k, data_plane_cidr)
		if left > 0 {
			ips = append(ips, ip_gen_sub(left, k+1, data_plane_cidr)...)
		}
	} else {
		ips = ip_gen_sub(count, k, data_plane_cidr)
		for iter := 1; iter <= t; iter++ {
			ips = append(ips, ip_gen_sub(count, k+iter, data_plane_cidr)...)
		}
		left = ip_num - t*count
		if left > 0 {
			ips = append(ips, ip_gen_sub(left, k+t+1, data_plane_cidr)...)
		}
	}

	return ips
}

func Node_port_gen(intf_num int, dev_list []string, dev_type string, ips []string, ip_flag bool) ([]database.Vport, []string) {
	// var nodes []database.Vnode
	var node database.Vnode
	var nics []database.Nic
	var nic database.Nic
	var ports []database.Vport
	var port database.Vport

	for _, dev := range dev_list {
		dev_intf := Intf_name(intf_num, dev)

		for _, dev_int := range dev_intf {
			nic.Intf = dev_int
			port.Intf = dev_int
			port.Id = GenUUID()
			port.Name = "vport" + ":" + port.Id
			if ip_flag {
				nic.Ip, ips = ips[len(ips)-1], ips[:len(ips)-1]
				port.Ip = nic.Ip
			} else {
				nic.Ip = ""
				port.Ip = ""
			}

			nics = append(nics, nic)
			ports = append(ports, port)
		}

		node.Id = GenUUID()
		node.Name = dev + ":" + node.Id
		node.Nics = nics
		Topo_nodes = append(Topo_nodes, node)
	}
	return ports, ips
}

// paring logic-- len(src) > len(dst), 1-to-1 paired, return the left dst
func Link_gen(ports1 []database.Vport, ports2 []database.Vport) []database.Vport {
	// var paired_links []database.Vlink
	var link database.Vlink
	var src_ports []database.Vport
	var dst_ports []database.Vport

	if len(ports1) <= len(ports2) {
		src_ports = ports1
		dst_ports = ports2
	} else {
		src_ports = ports2
		dst_ports = ports1
	}

	for t := 0; t < len(src_ports); t++ {
		src := src_ports[len(src_ports)-1]
		src_ports = src_ports[:len(src_ports)-1]
		dst := dst_ports[len(dst_ports)-1]
		dst_ports = dst_ports[:len(dst_ports)-1]

		link.Id = GenUUID()
		link.Name = "vlink:" + link.Id
		link.Src = src
		link.Dst = dst
		// paired_links = append(paired_links, link)
		Topo_links = append(Topo_links, link)
	}

	return dst_ports
}
