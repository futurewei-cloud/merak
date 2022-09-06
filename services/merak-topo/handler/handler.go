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
	"context"
	"log"
	"regexp"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"

	"fmt"

	"strings"

	pb_common "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//function CREATE
/* save the part of gw creation and mac learning for future requirment, comment the related code now*/
func Create(k8client *kubernetes.Clientset, topo_id string, aca_num uint32, rack_num uint32, aca_per_rack uint32, cgw_num uint32, data_plane_cidr string, ports_per_vswitch uint32, returnMessage *pb.ReturnTopologyMessage) error {

	log.Println("=== Parse gRPC message === ")
	log.Printf("Vhost number is: %v\n", aca_num)
	log.Printf("Rack number is: %v\n", rack_num)
	log.Printf("Vhosts per rack is: %v\n", aca_per_rack)

	log.Printf("Ports per vswitch is: %v\n", ports_per_vswitch)

	log.Println("=== Generate topology data === ")

	err_create, topo := Create_multiple_layers_vswitches(int(aca_num), int(rack_num), int(aca_per_rack), int(ports_per_vswitch), data_plane_cidr)
	if err_create != nil {
		return fmt.Errorf("create multiple layers vswitches error %s", err_create)
	}

	topo.Topology_id = topo_id

	log.Println("=== Save topology to redis ===")
	err1 := Topo_save(topo)
	if err1 != nil {
		return fmt.Errorf("save topology to redis error %s", err1)
	}

	log.Printf("topology details:  %v", topo)

	log.Println("=== Get k8s host nodes information ===")

	k8s_nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err1 != nil {
		return fmt.Errorf("failed to list k8s host nodes info %s", err1)
	}

	for _, s := range k8s_nodes.Items {
		var hnode pb_common.InternalHostInfo

		node_yaml, err2 := k8client.CoreV1().Nodes().Get(Ctx, s.Name, metav1.GetOptions{})
		if err2 != nil {
			return fmt.Errorf("failed to get k8s host node info %s", err2)
		}

		for _, c := range s.Status.Conditions {
			if c.Type == corev1.NodeReady {
				hnode.Status = pb_common.Status_READY
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

	log.Println("=== Return deployed compute nodes information ===")

	for _, node := range topo.Vnodes {
		var cnode pb_common.InternalComputeInfo
		if strings.Contains(node.Name, "vhost") {
			cnode.Name = node.Name
			cnode.Id = node.Id
			cnode.DatapathIp = strings.Split(node.Nics[len(node.Nics)-1].Ip, "/")[0]
			cnode.Veth = node.Nics[len(node.Nics)-1].Intf
		}

		log.Println("get compute nodes IP and Veth")

		cnode.OperationType = pb_common.OperationType_CREATE
		cnode.Status = pb_common.Status_DEPLOYING

		log.Println("get compute nodes status")

		returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &cnode)

		log.Println("generate returnMessage for compute nodes")

	}

	err_db2 := database.SetPbReturnValue(topo_id+":initialreturn", returnMessage)
	if err_db2 != nil {

		return fmt.Errorf("fail to save return msg to DB %s", err_db2)
	}

	log.Println("=== Topology Deployment === ")

	err_deploy := Topo_deploy(k8client, topo)

	if err_deploy != nil {
		return fmt.Errorf("topology deployment error %s", err_deploy)
	}

	return nil

}

// func Create(k8client *kubernetes.Clientset, topo_id string, aca_num uint32, rack_num uint32, aca_per_rack uint32, cgw_num uint32, data_plane_cidr string, returnMessage *pb.ReturnTopologyMessage) error {

// 	var topo database.TopologyData

// 	var ovs_tor_device = []string{"core-0"}
// 	ip_num := int(aca_num) + int(cgw_num)

// 	log.Println("=== parse done === ")
// 	log.Printf("Core switch is: %v \n", ovs_tor_device[0])
// 	log.Printf("Vswitch number is: %v\n", rack_num)
// 	log.Printf("Vhost number is: %v\n", aca_num)

// 	fmt.Println("======== Generate device list ==== ")
// 	rack_device := Pod_name(int(rack_num), "vswitch")
// 	aca_device := Pod_name(int(aca_num), "vhost")
// 	// ngw_device := Pod_name(int(cgw_num), "cgw")    /*comment gw creation function*/

// 	fmt.Printf("Vswitch_device: %v\n", rack_device)
// 	fmt.Printf("Vhost_device: %v\n", aca_device)
// 	// fmt.Printf("Cgw_device: %v\n", ngw_device) /*comment gw creation function*/

// 	fmt.Println("======== Generate device nodes ==== ")
// 	rack_intf_num := int(aca_per_rack) + 1
// 	tor_intf_num := int(rack_num) + int(cgw_num)
// 	aca_intf_num := 1
// 	// ngw_intf_num := 1 /*comment gw creation function*/

// 	log.Println("=== Generate ip addresses == ")

// 	ips := Ips_gen(topo_id, ip_num, data_plane_cidr)

// 	err := database.SetValue(topo_id+":ips", ips)
// 	if err != nil {
// 		return fmt.Errorf("fail to save ips in DB %s", err)
// 	}

// 	fmt.Println("======== Generate topology data ==== ")

// 	topo.Topology_id = topo_id

// 	topo_nodes, _ := Node_port_gen(aca_intf_num, aca_device, ips, true)
// 	// nodes, _ := Node_port_gen(ngw_intf_num, ngw_device, ips_1, true)  /*comment gw creation function*/
// 	// topo_nodes = append(topo_nodes, nodes...)
// 	nodes_s, _ := Node_port_gen(rack_intf_num, rack_device, ips, false)
// 	topo_nodes = append(topo_nodes, nodes_s...)
// 	nodes_t, _ := Node_port_gen(tor_intf_num, ovs_tor_device, ips, false)
// 	topo_nodes = append(topo_nodes, nodes_t...)

// 	fmt.Printf("The topology nodes are : %+v. \n", topo_nodes)

// 	topo.Vnodes = topo_nodes

// 	fmt.Println("======== Pairing links ==== ")

// 	topo_links := Links_gen(topo_nodes)

// 	fmt.Printf("The topology links are : %v. \n", topo_links)

// 	fmt.Printf("The topology total links are : %v. \n", len(topo_links))
// 	topo.Vlinks = topo_links

// 	fmt.Println("======== Save topo to redis =====")
// 	err1 := Topo_save(topo)
// 	if err1 != nil {
// 		return fmt.Errorf("save topo to redis error %s", err1)
// 	}

// 	fmt.Println("======== Topology Deployment ==== ")

// 	err_deploy := Topo_deploy(k8client, topo)

// 	if err_deploy != nil {
// 		return fmt.Errorf("topology deployment error %s", err_deploy)
// 	}

// 	fmt.Println("========= Get k8s host nodes information after deployment=====")

// 	k8s_nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

// 	if err1 != nil {
// 		return fmt.Errorf("failed to list k8s host nodes info %s", err1)
// 	}

// 	for _, s := range k8s_nodes.Items {
// 		var hnode pb_common.InternalHostInfo

// 		node_yaml, err2 := k8client.CoreV1().Nodes().Get(Ctx, s.Name, metav1.GetOptions{})
// 		if err2 != nil {
// 			return fmt.Errorf("failed to get k8s host node info %s", err2)
// 		}

// 		for _, c := range s.Status.Conditions {
// 			if c.Type == corev1.NodeReady {
// 				hnode.Status = pb_common.Status_READY
// 				log.Printf(s.Name + " status " + string(c.Type))
// 				break
// 			}

// 		}

// 		for _, res := range node_yaml.Status.Addresses {
// 			if res.Type == "InternalIP" {
// 				hnode.Ip = res.Address
// 				log.Printf(s.Name + " InternalIP " + string(hnode.Ip))
// 				break
// 			}
// 		}

// 		returnMessage.Hosts = append(returnMessage.Hosts, &hnode)

// 	}

// 	fmt.Println("========= Return deployed compute nodes information =====")

// 	for _, node := range topo.Vnodes {
// 		var cnode pb_common.InternalComputeInfo
// 		if strings.Contains(node.Name, "vhost") {
// 			cnode.Name = node.Name
// 			cnode.Id = node.Id
// 			cnode.DatapathIp = strings.Split(node.Nics[len(node.Nics)-1].Ip, "/")[0]
// 			cnode.Veth = node.Nics[len(node.Nics)-1].Intf
// 		}

// 		log.Printf("get compute nodes IP and Veth")

// 		cnode.OperationType = pb_common.OperationType_CREATE
// 		cnode.Status = pb_common.Status_DEPLOYING

// 		log.Printf("get compute nodes status")

// 		returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &cnode)

// 		log.Printf("generate returnMessage for compute nodes")

// 	}

// 	err_db2 := database.SetPbReturnValue(topo_id+":initialreturn", returnMessage)
// 	if err_db2 != nil {
// 		log.Printf("fail to save return msg to DB %s", err_db2.Error())
// 		return fmt.Errorf("fail to save return msg to DB %s", err_db2)
// 	}

// 	return nil
// }

func UpdateComputenodeInfo(client *kubernetes.Clientset, topo_id string, returnMessage *pb.ReturnTopologyMessage) error {
	topo, err := database.FindTopoEntity(topo_id, "")

	if err != nil {
		return fmt.Errorf("query topology_id error %s", err)
	}

	log.Printf("updatecomputenode:=========Get topo based on topo_id ===========")

	k8s_nodes, err1 := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err1 != nil {
		return fmt.Errorf("failed to list k8s host nodes info %s", err1)
	}

	for _, s := range k8s_nodes.Items {
		var hnode pb_common.InternalHostInfo

		node_yaml, err2 := client.CoreV1().Nodes().Get(Ctx, s.Name, metav1.GetOptions{})
		if err2 != nil {
			return fmt.Errorf("failed to get k8s host node info %s", err2)
		}

		for _, c := range s.Status.Conditions {
			if c.Type == corev1.NodeReady {
				hnode.Status = pb_common.Status_READY
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

	/*comment mac learning function*/

	// err3 := QueryMac(client, topo_id)

	// if err3 != nil {
	// 	return fmt.Errorf("query mac error %s", err3)
	// }

	// log.Printf("updatecomputenode:=========Update mac addresses ===========")

	for _, node := range topo.Vnodes {
		var cnode pb_common.InternalComputeInfo

		if strings.Contains(node.Name, "vhost") {

			res, err := client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})

			if err != nil {
				cnode.Name = node.Name
				cnode.Id = node.Id
				cnode.DatapathIp = strings.Split(node.Nics[len(node.Nics)-1].Ip, "/")[0]
				cnode.Veth = node.Nics[len(node.Nics)-1].Intf
				cnode.OperationType = pb_common.OperationType_INFO
				cnode.Status = pb_common.Status_ERROR

				returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &cnode)
				return fmt.Errorf("get pod error %s", err)
			} else {

				cnode.Name = res.Name
				cnode.Id = string(res.UID)

				cnode.DatapathIp = strings.Split(node.Nics[len(node.Nics)-1].Ip, "/")[0]
				cnode.Veth = node.Nics[len(node.Nics)-1].Intf
				if res.Status.PodIP != "" {
					cnode.ContainerIp = res.Status.PodIP
				} else {
					log.Printf("pod ip is not ready %v", res.Name)
				}

				/*comment mac learning function*/
				// mac, err := database.Get(topo_id + ":" + cnode.Ip)
				// if err != nil {
				// 	log.Printf("updatecomputenode: mac address is not available")
				// } else {
				// 	cnode.Mac = strings.Trim(mac, "\"")
				// 	cnode.OperationType = pb.OperationType_INFO
				// }

				cnode.Mac = "ff:ff:ff:ff:ff:ff"
				cnode.OperationType = pb_common.OperationType_INFO

				if len(res.Status.ContainerStatuses) == 0 {
					log.Printf("updatecomputenode: container status is not available %v", res.Name)
				} else {
					if res.Status.ContainerStatuses[len(res.Status.ContainerStatuses)-1].Ready {
						cnode.Status = pb_common.Status_READY
					} else {
						cnode.Status = pb_common.Status_DEPLOYING
					}
				}
				returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &cnode)
			}
		}
	}

	log.Printf("updatecomputenode:=========Update compute nodes information ===========")

	err_db2 := database.SetPbReturnValue(topo_id+":updateReturnmsg", returnMessage)
	if err_db2 != nil {
		log.Printf("fail to save k8s cluster host node to DB %s", err_db2)
		return fmt.Errorf("fail to save k8s cluster host node to DB %s", err_db2)
	}
	return nil
}

func Info(k8client *kubernetes.Clientset, topo_id string, returnMessage *pb.ReturnTopologyMessage) error {

	err := UpdateComputenodeInfo(k8client, topo_id, returnMessage)
	if err != nil {
		return fmt.Errorf("fail to update compute nodes info %s", err)
	}

	log.Printf("after updating compute node info: %s", returnMessage)

	err_check2 := database.GetPbReturnValue(topo_id+":updateReturnmsg", "", returnMessage)
	if err_check2 != nil {
		return fmt.Errorf("fail to find return message compute nodes info %s", err_check2)
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

func QueryMac(k8client *kubernetes.Clientset, topo_id string) error {

	topo_data, err := database.FindTopoEntity(topo_id, "")

	if err != nil {
		return fmt.Errorf("failed to retrieve topology data from DB %s", err)
	}
	count_ping := 0

	for _, node := range topo_data.Vnodes {

		if strings.Contains(node.Name, "cgw") && count_ping < 1 {
			count_ping = count_ping + 1

			pod, err := database.FindPodEntity(topo_id+":"+node.Name, "")

			if err != nil {
				return fmt.Errorf("failed to retrieve pod data from DB %s", err)
			}

			ip_list, err_db := database.FindIPEntity(topo_id+":ips", "")
			if err_db != nil {
				return fmt.Errorf("failed to retrieve ip address data from DB %s", err_db)
			}

			for index := len(ip_list) - 1; index >= 0; index-- {
				ip := ip_list[index]
				ip = strings.Split(ip, "/")[0]
				cmd1 := []string{
					"ping",
					"-c 1",
					ip,
				}
				Pod_query(k8client, pod, cmd1)

			}

			cmd2 := []string{
				"arp",
				"-a",
			}
			out2, err3 := Pod_query(k8client, pod, cmd2)

			if err3 != nil {
				return fmt.Errorf("failed to query pod compute node info from K8s %s", err)
			}

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
						database.SetValue(topo_id+":"+ip, mac)
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
