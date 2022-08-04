package handler

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
	"k8s.io/utils/strings/slices"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
)

// var (
// 	Topo_links []database.Vlink
// 	Topo_nodes []database.Vnode
// )

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
	} else if strings.Contains(prefix, "vswitch") {
		for i := 0; i < dev_num; i++ {
			intf_n = "vs" + strings.Split(prefix, "-")[1] + "-eth" + strconv.FormatInt(int64(i+1), 10)
			intf_name = append(intf_name, intf_n)
		}

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

func Ips_gen(topo_id string, ip_num int, data_plane_cidr string) []string {
	var count int = 250
	var ips = []string{}

	t := int(ip_num / count)

	if t == 0 {
		ips = ip_gen_sub(ip_num, 0, data_plane_cidr)
	} else {
		ips = ip_gen_sub(count, 0, data_plane_cidr)
		for iter := 1; iter < t; iter++ {
			ips = append(ips, ip_gen_sub(count, iter, data_plane_cidr)...)
		}
		left := ip_num - t*count
		if left > 0 {
			ips = append(ips, ip_gen_sub(left, t, data_plane_cidr)...)
		}
	}

	return ips
}

func Node_port_gen(intf_num int, dev_list []string, ips []string, ip_flag bool) ([]database.Vnode, []string) {

	var port database.Vport
	var nodes []database.Vnode

	for _, dev := range dev_list {
		var node database.Vnode
		var nics []database.Nic
		var nic database.Nic
		dev_intf := Intf_name(intf_num, dev)

		for _, dev_int := range dev_intf {
			nic.Intf = dev_int
			port.Intf = dev_int
			port.Id = GenUUID()
			port.Name = "vport" + "-" + dev + "-" + port.Id
			if ip_flag {
				nic.Ip, ips = ips[len(ips)-1], ips[:len(ips)-1]
				port.Ip = nic.Ip
			} else {
				nic.Ip = ""
				port.Ip = ""
			}

			nics = append(nics, nic)
			// ports = append(ports, port)
		}

		node.Id = GenUUID()
		// node.Name = dev + ":" + node.Id
		// node.Name = strings.Split(dev, "-")[0] + strings.Split(node.Id, "-")[0] + strings.Split(dev, "-")[1]
		node.Name = dev
		node.Nics = nics
		nodes = append(nodes, node)
	}
	return nodes, ips
}

func config_sclink(link database.Vlink) map[string]interface{} {

	config_clink := map[string]interface{}{
		"peer_pod":   strings.Split(link.Name, ":")[2],
		"local_intf": link.Src.Intf,
		"local_ip":   link.Src.Ip,
		"peer_intf":  link.Dst.Intf,
		"peer_ip":    link.Dst.Ip,
	}

	return config_clink
}

func config_dclink(link database.Vlink) map[string]interface{} {

	config_clink := map[string]interface{}{
		"peer_pod":   strings.Split(link.Name, ":")[1],
		"local_intf": link.Dst.Intf,
		"local_ip":   link.Dst.Ip,
		"peer_intf":  link.Src.Intf,
		"peer_ip":    link.Src.Ip,
	}

	return config_clink
}

func link_gen(src_name string, dst_name string, snic database.Nic, dnic database.Nic) database.Vlink {
	var link database.Vlink
	var link_dst database.Vport
	var link_src database.Vport

	link_dst.Id = GenUUID()
	link_dst.Name = "vport" + ":" + dst_name + ":" + link_dst.Id
	link_dst.Intf = dnic.Intf
	link_dst.Ip = dnic.Ip

	link_src.Id = GenUUID()
	link_src.Name = "vport" + ":" + src_name + ":" + link_src.Id
	link_src.Intf = snic.Intf
	link_src.Ip = snic.Ip

	link.Id = GenUUID()
	link.Name = "vlink" + ":" + src_name + ":" + dst_name
	link.Src = link_src
	link.Dst = link_dst

	return link

}

func Links_gen(topo_nodes []database.Vnode) []database.Vlink {
	src_nodes := topo_nodes
	dst_nodes := topo_nodes
	var topo_links []database.Vlink

	picked_intf := []string{}

	for i, s := range src_nodes {
		node_name := strings.Split(s.Name, ":")[0]

		if strings.Contains(node_name, "core") {

			var paired_nodes []string
			for _, snic := range s.Nics {

				var paired = false

				// fmt.Printf("===snic %v===\n", snic.Intf)

				if !slices.Contains(picked_intf, snic.Intf) && !paired {
					picked_intf = append(picked_intf, snic.Intf)

					for j, d := range dst_nodes {

						dst_name := strings.Split(d.Name, ":")[0]

						if (strings.Contains(dst_name, "ovs")) && (!slices.Contains(paired_nodes, dst_name)) && !paired {
							paired_nodes = append(paired_nodes, dst_name)
							for _, dnic := range d.Nics {
								if !slices.Contains(picked_intf, dnic.Intf) && !paired {
									picked_intf = append(picked_intf, dnic.Intf)
									paired = true
									// fmt.Printf("==dst Intf == %v \n", dnic.Intf)
									link := link_gen(node_name, dst_name, snic, dnic)
									topo_links = append(topo_links, link)

									s_clink := config_sclink(link)
									s_clink["uid"] = len(topo_links)
									topo_nodes[i].Flinks = append(topo_nodes[i].Flinks, s_clink)

									d_clink := config_dclink(link)
									d_clink["uid"] = len(topo_links)
									topo_nodes[j].Flinks = append(topo_nodes[j].Flinks, d_clink)

								}
							}
						}
					}

				}

			}

		}

	}

	for i, s := range src_nodes {
		node_name := strings.Split(s.Name, ":")[0]

		if strings.Contains(node_name, "ovs") {

			var paired_nodes []string
			for _, snic := range s.Nics {

				var paired = false

				// fmt.Printf("===snic %v===\n", snic.Intf)

				if !slices.Contains(picked_intf, snic.Intf) && !paired {
					picked_intf = append(picked_intf, snic.Intf)

					for j, d := range dst_nodes {

						dst_name := strings.Split(d.Name, ":")[0]

						if strings.Contains(dst_name, "vswitch") && (!slices.Contains(paired_nodes, dst_name)) && !paired {
							paired_nodes = append(paired_nodes, dst_name)
							for _, dnic := range d.Nics {
								if !slices.Contains(picked_intf, dnic.Intf) && !paired {
									picked_intf = append(picked_intf, dnic.Intf)
									paired = true
									// fmt.Printf("==dst Intf == %v \n", dnic.Intf)
									link := link_gen(node_name, dst_name, snic, dnic)
									topo_links = append(topo_links, link)

									s_clink := config_sclink(link)
									s_clink["uid"] = len(topo_links)
									topo_nodes[i].Flinks = append(topo_nodes[i].Flinks, s_clink)

									d_clink := config_dclink(link)
									d_clink["uid"] = len(topo_links)
									topo_nodes[j].Flinks = append(topo_nodes[j].Flinks, d_clink)

								}
							}
						}
					}

				}

			}

		}

	}

	for i, s := range src_nodes {
		node_name := strings.Split(s.Name, ":")[0]
		if strings.Contains(node_name, "vhost") {

			var paired_nodes []string

			for _, snic := range s.Nics {
				// fmt.Printf("===snic %v===\n", snic.Intf)
				var paired = false
				if !slices.Contains(picked_intf, snic.Intf) {
					picked_intf = append(picked_intf, snic.Intf)

					for j, d := range dst_nodes {

						dst_name := strings.Split(d.Name, ":")[0]
						if (strings.Contains(dst_name, "vswitch")) && (!slices.Contains(paired_nodes, dst_name)) && !paired {

							paired_nodes = append(paired_nodes, dst_name)

							for _, dnic := range d.Nics {
								if !slices.Contains(picked_intf, dnic.Intf) && !paired {
									picked_intf = append(picked_intf, dnic.Intf)
									// fmt.Printf("==dst Intf == %v \n", dnic.Intf)
									paired = true
									link := link_gen(node_name, dst_name, snic, dnic)
									topo_links = append(topo_links, link)

									s_clink := config_sclink(link)
									s_clink["uid"] = len(topo_links)
									topo_nodes[i].Flinks = append(topo_nodes[i].Flinks, s_clink)

									d_clink := config_dclink(link)
									d_clink["uid"] = len(topo_links)
									topo_nodes[j].Flinks = append(topo_nodes[j].Flinks, d_clink)

								}
							}
						}
					}
				}

			}

		}

	}

	for i, s := range src_nodes {
		node_name := strings.Split(s.Name, ":")[0]
		if strings.Contains(node_name, "cgw") {

			var paired_nodes []string

			for _, snic := range s.Nics {
				// fmt.Printf("===snic %v===\n", snic.Intf)
				var paired = false
				if !slices.Contains(picked_intf, snic.Intf) {
					picked_intf = append(picked_intf, snic.Intf)

					for j, d := range dst_nodes {

						dst_name := strings.Split(d.Name, ":")[0]
						if (strings.Contains(dst_name, "vswitch")) && (!slices.Contains(paired_nodes, dst_name)) && !paired {

							paired_nodes = append(paired_nodes, dst_name)

							for _, dnic := range d.Nics {
								if !slices.Contains(picked_intf, dnic.Intf) && !paired {
									picked_intf = append(picked_intf, dnic.Intf)
									// fmt.Printf("==dst Intf == %v \n", dnic.Intf)
									paired = true
									link := link_gen(node_name, dst_name, snic, dnic)
									topo_links = append(topo_links, link)

									s_clink := config_sclink(link)
									s_clink["uid"] = len(topo_links)
									topo_nodes[i].Flinks = append(topo_nodes[i].Flinks, s_clink)

									d_clink := config_dclink(link)
									d_clink["uid"] = len(topo_links)
									topo_nodes[j].Flinks = append(topo_nodes[j].Flinks, d_clink)

								}
							}
						}
					}
				}

			}

		}

	}
	return topo_links

}
