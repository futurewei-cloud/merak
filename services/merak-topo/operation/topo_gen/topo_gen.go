package operation

import (
	"strconv"
	"strings"
)

//function methods

func podName(num int, prefix string) []string {
	var pod_name = []string{}
	var pod_n string = ""
	var i int = 0

	if prefix == "ovs" {
		for i < num {
			pod_n = prefix + "-" + strconv.FormatInt(int64(i+1), 10)
			i = i + 1
			pod_name = append(pod_name, pod_n)
		}
	} else {
		for i < num {
			pod_n = prefix + "-" + strconv.FormatInt(int64(i), 10)
			i = i + 1
			pod_name = append(pod_name, pod_n)
		}
	}
	return pod_name
}

func intfName(dev_num int, prefix string) []string {
	var intf_name = []string{}
	var i int = 0
	var intf_n string = ""
	if strings.Contains(prefix, "ovs-0") {
		intf_num := dev_num
		for i < intf_num {
			intf_n = strings.Split(prefix, "-")[0] + strings.Split(prefix, "-")[1] + "-eth" + strconv.FormatInt(int64(i+1), 10)
			i = i + 1
			intf_name = append(intf_name, intf_n)
		}
	} else if strings.Contains(prefix, "aca") {
		intf_n = strings.Split(prefix, "-")[0] + strings.Split(prefix, "-")[1] + "-eth1"
		intf_name = append(intf_name, intf_n)
	} else {
		intf_num := dev_num
		for i < intf_num {
			intf_n = strings.Split(prefix, "-")[0] + strings.Split(prefix, "-")[1] + "-eth" + strconv.FormatInt(int64(i+2), 10)
			i = i + 1
			intf_name = append(intf_name, intf_n)
		}

	}
	return intf_name
}

func ipGen(upper int, k1 int) []string {
	var ips_gen = []string{}
	var i int = 1
	var ip string = ""
	for i < upper+1 {
		ip = "10.200." + strconv.FormatInt(int64(k1), 10) + "." + strconv.FormatInt(int64(i+1), 10) + "/16"
		i = i + 1
		ips_gen = append(ips_gen, ip)
	}
	return ips_gen
}

func ipsPod(aca_num int, gw_num int, k int, count int) []string {
	var ips = []string{}
	var left = aca_num + gw_num - count
	var iter int = 1
	var t int = int((aca_num + gw_num) / count)

	if t == 0 {
		ips = ipGen((aca_num + gw_num), k)
	} else if t == 1 {
		ips = ipGen(count, k)
		if left > 0 {
			ips = append(ips, ipGen(left, k+1)...)
		}
	} else {
		ips = ipGen(count, k)
		for iter <= t {
			ips = append(ips, ipGen(count, k+iter)...)
			iter = iter + 1
		}
		left = aca_num + gw_num - t*count
		if left > 0 {
			ips = append(ips, ipGen(left, k+t+1)...)
		}
	}

	return ips
}

func link_pod_ip_single(dev_list []string, dev_type string, ips []string) []string {
	var pod_link_ip []string
	var pod_link_name string
	for i := 0; i < len(dev_list); i++ {
		dev_intf := intfName(1, dev_list[i])

		for j := 0; j < len(dev_intf); j++ {
			pod_link_name = dev_list[i] + ":" + dev_intf[j] + ":" + ips[len(ips)-1]
			ips = ips[:len(ips)-1]
			pod_link_ip = append(pod_link_ip, pod_link_name)
		}

	}

	return pod_link_ip
}

func link_pod(num int, dev_list []string) []string {
	var pod_link []string
	var pod_link_name string
	for i := 0; i < len(dev_list); i++ {
		dev_intf := intfName(1, dev_list[i])

		for j := 0; j < len(dev_intf); j++ {
			pod_link_name = dev_list[i] + ":" + dev_intf[j]

			pod_link = append(pod_link, pod_link_name)
		}

	}
	return pod_link
}

func ovs_tor_peers(dev_list []string, dev_type string, ips []string) []string {
	var ovs_tor_peers []string
	if dev_type == "gw" {
		for i := 0; i < len(dev_list); i++ {
			pod := dev_list[i]
			peer := pod + ":" + strings.Split(pod, "-")[0] + strings.Split(pod, "-")[1] + "-eth1" + ":" + ips[len(ips)-1]
			ips = ips[:len(ips)-1]
			ovs_tor_peers = append(ovs_tor_peers, peer)
		}
	} else {
		for i := 0; i < len(dev_list); i++ {
			pod := dev_list[i]
			peer := pod + ":" + strings.Split(pod, "-")[0] + strings.Split(pod, "-")[1] + "-eth1"

			ovs_tor_peers = append(ovs_tor_peers, peer)
		}
	}

	return ovs_tor_peers
}

func Topo_gen() {
	//
}
