package handler

import (
	"context"
	"log"
	"regexp"
	"strconv"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"

	"fmt"

	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	count int = 250
	k     int = 0 // subnet starting number
)

//function CREATE
func Create(k8client *kubernetes.Clientset, topo_id string, aca_num uint32, rack_num uint32, aca_per_rack uint32, cgw_num uint32, data_plane_cidr string, returnMessage *pb.ReturnTopologyMessage) error {

	var topo database.TopologyData

	var ovs_tor_device = []string{"tor-0"}
	ip_num := int(aca_num) + int(cgw_num)

	ips := Ips_gen(ip_num, k, count, data_plane_cidr)

	fmt.Println("=== parse done == ")
	fmt.Printf("TOR OVS is: %v \n", ovs_tor_device[0])
	fmt.Printf("Vswitch number is: %v\n", rack_num)
	fmt.Printf("Vhost number is: %v\n", aca_num)

	fmt.Println("======== Generate device list ==== ")
	rack_device := Pod_name(int(rack_num), "vswitch")
	aca_device := Pod_name(int(aca_num), "vhost")
	ngw_device := Pod_name(int(cgw_num), "cgw")

	fmt.Printf("Vswitch_device: %v\n", rack_device)
	fmt.Printf("Vhost_device: %v\n", aca_device)
	fmt.Printf("Cgw_device: %v\n", ngw_device)

	fmt.Println("======== Generate device nodes ==== ")
	rack_intf_num := int(aca_per_rack) + 1
	tor_intf_num := int(rack_num) + int(cgw_num)
	aca_intf_num := 1
	ngw_intf_num := 1

	ips_1 := Node_port_gen(aca_intf_num, aca_device, ips, true)
	Node_port_gen(ngw_intf_num, ngw_device, ips_1, true)
	Node_port_gen(rack_intf_num, rack_device, ips, false)
	Node_port_gen(tor_intf_num, ovs_tor_device, ips, false)

	fmt.Printf("The topology nodes are : %+v. \n", Topo_nodes)

	fmt.Println("======== Pairing links ==== ")

	Links_gen(Topo_nodes)
	// fmt.Printf("The topology links are : %+v. \n", Topo_links)

	fmt.Println("======== Generate topology data ==== ")
	topo.Topology_id = topo_id
	topo.Vlinks = Topo_links
	topo.Vnodes = Topo_nodes

	fmt.Println("======== Save topo to redis =====")
	err1 := Topo_save(topo)
	if err1 != nil {
		return fmt.Errorf("save topo to redis error %s", err1)
	}

	fmt.Println("======== Topology Deployment ==== ")

	err := Topo_deploy(k8client, topo)
	if err != nil {
		return fmt.Errorf("topology deployment error %s", err)
	}

	fmt.Println("========= Get k8s host nodes information after deployment=====")

	nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err1 != nil {
		return fmt.Errorf("failed to list k8s host nodes info %s", err1)
	}

	for _, s := range nodes.Items {
		var hnode pb.InternalHostInfo

		node_yaml, err2 := k8client.CoreV1().Nodes().Get(Ctx, s.Name, metav1.GetOptions{})
		if err2 != nil {
			return fmt.Errorf("failed to get k8s host node info %s", err2)
		}

		for _, c := range s.Status.Conditions {
			if c.Type == corev1.NodeReady {
				hnode.Status = pb.Status_READY
				break
			}
		}

		for _, res := range node_yaml.Status.Addresses {
			if res.Type == "InternalIP" {
				hnode.Ip = res.Address
				break
			}
		}

		returnMessage.Hosts = append(returnMessage.Hosts, &hnode)

	}

	fmt.Println("========= Get compute nodes information after deployment=====")

	// err3 := QueryMac(k8client, topo_id)

	// if err3 != nil {
	// 	return fmt.Errorf("query mac error %s", err3)
	// }

	for _, node := range topo.Vnodes {
		var cnode pb.InternalComputeInfo

		res, err := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})

		if err != nil {
			return fmt.Errorf("get pod error %s", err)
		} else {
			// if res.Status.Phase == "Running" && res.Labels["Type"] == "vhost" {
			if res.Labels["Type"] == "vhost" {
				cnode.Name = res.Name
				cnode.Id = string(res.UID)

				for _, n := range topo.Vnodes {
					if n.Name == res.Name {
						cnode.Ip = strings.Split(n.Nics[len(n.Nics)-1].Ip, "/")[0]
						cnode.Veth = n.Nics[len(n.Nics)-1].Intf
					}
				}

				// cnode.Mac, err = database.Get(topo_id + "-" + cnode.Ip)
				// if err != nil {
				// 	return fmt.Errorf("fail to get mac from db %s", err)
				// }

				cnode.OperationType = pb.OperationType_INFO
				if res.Status.Phase == "Running" {
					cnode.Status = pb.Status_READY
				} else {
					cnode.Status = pb.Status_DEPLOYING
				}

				returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &cnode)
			}
		}
	}

	return nil
}

func Info(k8client *kubernetes.Clientset, topo_id string, aca_num int, returnMessage *pb.ReturnTopologyMessage) error {

	topo, err := database.FindTopoEntity(topo_id, "")

	if err != nil {
		return fmt.Errorf("query topology_id error %s", err)
	}

	nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err1 != nil {
		return fmt.Errorf("failed to list k8s host nodes info %s", err1)
	}

	for _, s := range nodes.Items {
		var hnode pb.InternalHostInfo

		node_yaml, err2 := k8client.CoreV1().Nodes().Get(Ctx, s.Name, metav1.GetOptions{})
		if err2 != nil {
			return fmt.Errorf("failed to get k8s host node info %s", err2)
		}

		for _, c := range s.Status.Conditions {
			if c.Type == corev1.NodeReady {
				hnode.Status = pb.Status_READY
				log.Printf(s.Name + " status " + string(c.Type))
				break
			}

		}

		for _, res := range node_yaml.Status.Addresses {
			if res.Type == "InternalIP" {
				hnode.Ip = res.Address
				log.Printf(s.Name + " InternalIP " + string(hnode.Ip))
				break
			}
		}

		returnMessage.Hosts = append(returnMessage.Hosts, &hnode)

	}

	err3 := QueryMac(k8client, topo_id, aca_num)

	if err3 != nil {
		return fmt.Errorf("query mac error %s", err3)
	}

	for _, node := range topo.Vnodes {
		var cnode pb.InternalComputeInfo

		res, err := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})

		if err != nil {
			return fmt.Errorf("get pod error %s", err)
		} else {
			// if res.Status.Phase == "Running" && res.Labels["Type"] == "vhost" {
			if res.Labels["Type"] == "vhost" {
				cnode.Name = res.Name
				cnode.Id = string(res.UID)

				for _, n := range topo.Vnodes {
					if n.Name == res.Name {
						cnode.Ip = strings.Split(n.Nics[len(n.Nics)-1].Ip, "/")[0]
						cnode.Veth = n.Nics[len(n.Nics)-1].Intf
						break
					}
				}

				mac, err := database.Get(topo_id + "-" + cnode.Ip)
				cnode.Mac = strings.Trim(mac, "\"")
				// add DB structure to save mac, ip for a k8s cluster

				if err != nil {
					return fmt.Errorf("fail to get mac from db %s", err)
				}

				cnode.OperationType = pb.OperationType_INFO
				if res.Status.Phase == "Running" {
					cnode.Status = pb.Status_READY
				} else {
					cnode.Status = pb.Status_DEPLOYING
				}
				log.Printf(cnode.Name + " status " + string(cnode.Status))
				returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &cnode)
			}
		}
	}

	return nil
}

func Delete(k8client *kubernetes.Clientset, topo_id string) error {
	topo, err_db := database.FindTopoEntity(topo_id, "")

	if err_db != nil {
		return fmt.Errorf("query topology_id error %s", err_db)
	}

	err := Topo_delete(k8client, topo)
	if err != nil {
		return fmt.Errorf("topology delete fails %s", err)
	}

	return nil
}

func QueryMac(k8client *kubernetes.Clientset, topo_id string, aca_num int) error {

	topo_data, err := database.FindTopoEntity(topo_id, "")

	if err != nil {
		return fmt.Errorf("failed to retrieve topology data from DB %s", err)
	}
	count_ping := 0

	for _, node := range topo_data.Vnodes {

		// if strings.Contains(node.Name, "vhost") && count_ping < 2 {
		if strings.Contains(node.Name, "cgw") && count_ping < 1 {
			count_ping = count_ping + 1

			pod, err := database.FindPodEntity(topo_id+"-"+node.Name, "-pod")

			if err != nil {
				return fmt.Errorf("failed to retrieve pod data from DB %s", err)
			}

			var ip_last []string

			for aca_num > 0 {
				ip_last = append(ip_last, strconv.Itoa(aca_num))
				aca_num = aca_num - 1
			}

			for _, i := range ip_last {
				cmd1 := []string{
					"ping",
					"-c 1",
					"10.200.0." + i,
				}
				out1, _ := Pod_query(k8client, pod, cmd1)
				// if errping != nil {
				// 	return fmt.Errorf("failed to ping all %s", errping)
				// }
				log.Println(string(out1))

			}

			cmd2 := []string{
				"arp",
				"-a",
			}
			out2, err3 := Pod_query(k8client, pod, cmd2)

			if err3 != nil {
				return fmt.Errorf("failed to query pod compute node info from K8s %s", err)
			}

			log.Printf("arp output %v", out2)

			re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

			re2, _ := regexp.Compile(`^[a-fA-F0-9]{2}(:[a-fA-F0-9]{2}){5}$`)
			var ip string
			var mac string
			count := 0

			for _, s := range strings.Split(out2, " ") {
				s1 := strings.Trim(s, "(")
				s2 := strings.Trim(s1, ")")
				s2 = strings.Trim(s2, "\"")

				if re.MatchString(s2) {
					ip = s2
					count = count + 1
				} else if re2.MatchString(s) {
					mac = s2
					count = count + 1
					if count == 2 {
						database.SetValue(topo_id+"-"+ip, mac)
						count = 0
					}
				}
			}
			if count != 0 {
				log.Printf("arp output parsing error. ip and mac doesn't match")
			}
		}
	}
	return nil
}

func QueryHostNode(k8client *kubernetes.Clientset, topo_id string) error {

	nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err1 != nil {
		return fmt.Errorf("failed to retrieve topology data from DB %s", err1)
	}

	for _, s := range nodes.Items {
		node_yaml, err2 := k8client.CoreV1().Nodes().Get(Ctx, s.Name, metav1.GetOptions{})
		if err2 != nil {
			return fmt.Errorf("failed to retrieve topology data from DB %s", err2)
		}

		for _, c := range s.Status.Conditions {
			if c.Type == corev1.NodeReady {
				database.SetValue(topo_id+s.Name+"status", corev1.NodeReady)
				log.Printf(s.Name + "status" + string(c.Type))
				break
			} else {
				log.Printf(s.Name + "status" + string(c.Type))
			}
		}

		for _, res := range node_yaml.Status.Addresses {
			if res.Type == "InternalIP" {
				ip := res.Address
				database.SetValue(topo_id+s.Name+"ip", ip)
			}
		}
	}

	return nil
}

func Testapi(k8client *kubernetes.Clientset, topo database.TopologyData) ([]*pb.InternalComputeInfo, error) {
	var cnodes []*pb.InternalComputeInfo
	var cnode *pb.InternalComputeInfo

	for _, node := range topo.Vnodes {

		res, err := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})

		if err != nil {
			return cnodes, fmt.Errorf("get pod error %s", err)
		}
		if res.Status.Phase == "Running" && res.Labels["Type"] == "vhost" {
			cnode.Name = res.Name
			cnode.Id = string(res.UID)
			// cnode.HostIP = out.Items[i].Status.HostIP
			cnode.Ip = res.Status.PodIP
			cnode.Mac = ""
			cnode.Veth = ""
			cnode.OperationType = 2
			cnodes = append(cnodes, cnode)

		}
	}
	return cnodes, nil
}
