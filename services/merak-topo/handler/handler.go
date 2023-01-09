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
	"errors"

	"time"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"

	"strings"

	pb_common "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//function CREATE
/* save the part of gw creation and mac learning for future requirment, comment the related code now*/

func Create(k8client *kubernetes.Clientset, topo_id string, aca_num uint32, rack_num uint32, aca_per_rack uint32, cgw_num uint32, data_plane_cidr string, ports_per_vswitch uint32, images []*pb.InternalTopologyImage, aca_parameters string, returnMessage *pb.ReturnTopologyMessage) error {

	start_time := time.Now()
	err_flag := 0
	errs := errors.New("request DEPLOY- fails to create topology")

	var aca_image string
	var ovs_image string
	var namespace string

	for _, img := range images {
		if strings.Contains(img.Name, "ACA") {
			aca_image = img.Registry
		} else if strings.Contains(img.Name, "OVS") {
			ovs_image = img.Registry
		}
	}

	var err_return error

	utils.Logger.Debug("request DEPLOY details", "Vhost number", aca_num, "Rack number", rack_num, "Vhosts per rack", aca_per_rack, "Ports per vswitch", ports_per_vswitch)

	topo, err_create := Create_multiple_layers_vswitches(int(aca_num), int(rack_num), int(aca_per_rack), int(ports_per_vswitch), data_plane_cidr)
	if err_create != nil {
		utils.Logger.Error("request DEPLOY", "multiple layers vswitches", err_create.Error())
		returnMessage.ReturnMessage = "Can not create multiple layers vswitches"
		returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
		err_flag = 1

	}

	topo.Topology_id = topo_id

	elaps0 := time.Since(start_time)
	start0 := time.Now()

	utils.Logger.Info("request DEPLOY", " Complete: Generate topology data (in second)", elaps0)

	err1 := Topo_save(topo)
	if err1 != nil {
		utils.Logger.Error("request DEPLOY", "save topology to redis", err1.Error(), "topo_id", topo.Topology_id)
		returnMessage.ReturnMessage = "Can not save topology to redis"
		returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
		err_flag = 1
	}

	elaps1 := time.Since(start0)
	start1 := time.Now()
	utils.Logger.Info("request DEPLOY", "Save topology in redis DB (in second)", elaps1)

	k8s_nodes, err2 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err2 != nil {
		utils.Logger.Error("request DEPLOY", "check nodes in k8s cluster", err2.Error())
		returnMessage.ReturnMessage = "CREATE: can not list k8s host nodes info"
		returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
		err_flag = 1
	}

	for _, s := range k8s_nodes.Items {
		var hnode database.HostNode

		node_yaml, err3 := k8client.CoreV1().Nodes().Get(Ctx, s.Name, metav1.GetOptions{})
		if err3 != nil {
			utils.Logger.Error("can't get host node info in k8s cluster", s.Name, err3.Error())
			returnMessage.ReturnMessage = "DEPLOY: can not get k8s host node info from k8s cluster"
			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			err_flag = 1
		}

		for _, c := range s.Status.Conditions {
			if c.Type == corev1.NodeReady {
				hnode.Status = database.STATUS_READY
				break
			}

		}

		for _, res := range node_yaml.Status.Addresses {
			if res.Type == "InternalIP" {
				hnode.Ip = res.Address
				break
			}
		}

		// Make return message with k8s cluster nodes info
		err := database.SetValue(topo_id[:5]+":"+s.Name, hnode)
		if err != nil {
			utils.Logger.Error("can not save host node info in DB", "key", topo_id+":"+s.Name, "error msg", err.Error())
			returnMessage.ReturnMessage = "DEPLOY: can not save host node info in DB"
			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			err_flag = 1
		}

	}

	for _, s := range k8s_nodes.Items {
		var hnode database.HostNode
		var hrm pb_common.InternalHostInfo

		h, err := database.FindHostEntity(topo_id[:5]+":"+s.Name, "")

		hnode = h

		if err != nil {
			utils.Logger.Error("cannot query host node info from DB ", topo_id+":"+s.Name, err.Error())
			returnMessage.ReturnMessage = "DEPLOY: cannot query host node info from DB"
			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			err_flag = 1
		} else {
			hrm.Ip = hnode.Ip
			if hnode.Status == database.STATUS_READY {
				hrm.Status = pb_common.Status_READY

			} else {
				hrm.Status = pb_common.Status_NONE
			}

			hrm.RoutingRules = hnode.Routing_rule

			returnMessage.Hosts = append(returnMessage.Hosts, &hrm)
		}

	}

	for _, node := range topo.Vnodes {
		var cnode database.ComputeNode
		var crm pb_common.InternalComputeInfo
		if strings.Contains(node.Name, "vhost") {
			cnode.Name = node.Name
			cnode.Id = node.Id
			cnode.DatapathIp = strings.Split(node.Nics[len(node.Nics)-1].Ip, "/")[0]
			cnode.Veth = node.Nics[len(node.Nics)-1].Intf

			cnode.OperationType = database.OPERATION_CREATE
			cnode.Mac = "ff:ff:ff:ff:ff:ff"
			cnode.Status = database.STATUS_DEPLOYING

			err := database.SetValue(topo_id[:5]+":"+cnode.Name, cnode)
			if err != nil {
				utils.Logger.Error("can not save compute node info in DB ", topo_id[:5]+":"+cnode.Name, err.Error())
				returnMessage.ReturnMessage = "DEPLOY: can not save compute node info in DB"
				returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
				err_flag = 1
			}
			crm.Id = cnode.Id
			crm.Name = cnode.Name
			crm.DatapathIp = cnode.DatapathIp
			crm.Mac = cnode.Mac
			crm.Veth = cnode.Veth
			crm.Status = pb_common.Status_DEPLOYING
			crm.OperationType = pb_common.OperationType_CREATE
			returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &crm)

		}

	}

	// for _, node := range topo.Vnodes {
	// 	if strings.Contains(node.Name, "vhost") {
	// 		var cnode database.ComputeNode
	// 		var crm pb_common.InternalComputeInfo

	// 		c, err := database.FindComputeEntity(topo_id[:5]+":"+node.Name, "")

	// 		cnode = c

	// 		if err != nil {
	// 			utils.Logger.Error("can't get compute node from DB", topo_id+":"+node.Name, err.Error())
	// 			err_flag = 1
	// 		}

	// 		crm.Id = cnode.Id
	// 		crm.Name = cnode.Name
	// 		crm.DatapathIp = cnode.DatapathIp
	// 		crm.ContainerIp = cnode.ContainerIp
	// 		crm.Mac = cnode.Mac
	// 		crm.Veth = cnode.Veth
	// 		if cnode.Status == database.STATUS_READY {
	// 			crm.Status = pb_common.Status_READY
	// 		} else if cnode.Status == database.STATUS_DELETING {
	// 			crm.Status = pb_common.Status_DELETING
	// 		} else if cnode.Status == database.STATUS_DEPLOYING {
	// 			crm.Status = pb_common.Status_DEPLOYING
	// 		} else {
	// 			crm.Status = pb_common.Status_NONE
	// 		}
	// 		if cnode.OperationType == database.OPERATION_CREATE {
	// 			crm.OperationType = pb_common.OperationType\_CREATE
	// 		} else if cnode.OperationType == database.OPERATION_INFO {
	// 			crm.OperationType = pb_common.OperationType_INFO
	// 		} else if cnode.OperationType == database.OPERATION_DELETE {
	// 			crm.OperationType = pb_common.OperationType_DELETE
	// 		} else if cnode.OperationType == database.OPERATION_UPDATE {
	// 			crm.OperationType = pb_common.OperationType_UPDATE
	// 		}

	// 		returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &crm)

	// 	}

	// }

	elaps2 := time.Since(start1)

	utils.Logger.Info("request DEPLOY", "Complete: Return compute nodes and k8s cluster information to Scenario Manager (in second)", elaps2)

	namespace = "merak-" + topo_id[:5]

	nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}

	k8client.CoreV1().Namespaces().Create(context.Background(), nsSpec, metav1.CreateOptions{})

	go Topo_deploy(k8client, aca_image, ovs_image, topo, aca_parameters, namespace)
	// Topo_deploy(k8client, aca_image, ovs_image, topo, aca_parameters, namespace)

	if err_flag == 1 {
		err_return = errs
	} else {
		err_return = nil
	}

	return err_return

}

func UpdateComputenodeInfo(k8client *kubernetes.Clientset, topo_id string, namespace string) error {
	err_flag := 0
	errs := errors.New("fails to update compute nodes and k8s node info in DB")
	var err_return error

	start_time := time.Now()
	topo, err := database.FindTopoEntity(topo_id[:5], "")
	if err != nil {
		utils.Logger.Error("can't find topology info in DB", topo_id, err.Error())
		err_flag = 1
	}

	k8s_nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err1 != nil {
		utils.Logger.Error("can't check k8s cluster ", "corev1", err1.Error())
		err_flag = 1
	}

	for _, s := range k8s_nodes.Items {
		var hnode database.HostNode

		node_yaml, err2 := k8client.CoreV1().Nodes().Get(Ctx, s.Name, metav1.GetOptions{})
		if err2 != nil {
			utils.Logger.Error("can't get k8s host node info ", s.Name, err2.Error())
			err_flag = 1
		}

		for _, c := range s.Status.Conditions {
			if c.Type == corev1.NodeReady {
				hnode.Status = database.STATUS_READY
				break
			}

		}

		for _, res := range node_yaml.Status.Addresses {
			if res.Type == "InternalIP" {
				hnode.Ip = res.Address
				break
			}
		}

		// Make return message with k8s cluster nodes info
		err := database.SetValue(topo_id[:5]+":"+s.Name, hnode)
		if err != nil {
			utils.Logger.Error("can't save host node in DB", topo_id[:5]+":"+s.Name, err.Error())
			err_flag = 1
		}

	}

	elaps0 := time.Since(start_time)
	start_time0 := time.Now()

	utils.Logger.Info("Complete", "updating host node info in DB (in second) ", elaps0)

	/*comment mac learning function*/

	// err3 := QueryMac(client, topo_id)

	// if err3 != nil {
	// 	return fmt.Errorf("query mac error %s", err3)
	// }

	// log.Printf("updatecomputenode:=Update mac addresses =")

	for _, node := range topo.Vnodes {

		if strings.Contains(node.Name, "vhost") {

			var cnode database.ComputeNode

			res, err := k8client.CoreV1().Pods(namespace).Get(Ctx, node.Name, metav1.GetOptions{})

			if err != nil {
				utils.Logger.Error("can't get pod info from k8s", node.Name, err.Error(), "namespace", namespace)
				err_flag = 1
			} else {
				cnode.Name = res.Name
				cnode.Id = string(res.UID)
				cnode.DatapathIp = strings.Split(node.Nics[len(node.Nics)-1].Ip, "/")[0]
				cnode.Veth = node.Nics[len(node.Nics)-1].Intf
				if res.Status.PodIP != "" {
					cnode.ContainerIp = res.Status.PodIP
				} else {
					utils.Logger.Debug("Warning", "pod ip is not ready", res.Name)
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
				cnode.OperationType = database.OPERATION_INFO

				if len(res.Status.ContainerStatuses) == 0 {
					cnode.Status = database.STATUS_NONE
					utils.Logger.Debug("Warning", "container status is not available ", res.Name)
				} else {
					if res.Status.ContainerStatuses[len(res.Status.ContainerStatuses)-1].Ready {
						cnode.Status = database.STATUS_READY
					}

				}

				err_db := database.SetValue(topo_id[:5]+":"+node.Name, cnode)
				if err_db != nil {
					utils.Logger.Error("can't save compute node in DB", topo_id[:5]+":"+node.Name, err_db.Error())
					err_flag = 1
				}

			}

		}

	}

	elaps := time.Since(start_time0)

	utils.Logger.Info("Complete", "updating compute nodes info in DB (in second) ", elaps)

	if err_flag == 1 {
		err_return = errs
		err_flag = 0
	} else {
		err_return = nil
	}
	return err_return
}

func Info(k8client *kubernetes.Clientset, topo_id string, returnMessage *pb.ReturnTopologyMessage) error {

	err_flag := 0
	errs := errors.New("fails to handle request CHECK")
	var err_return error

	namespace := "merak-" + topo_id[:5]

	topo, err := database.FindTopoEntity(topo_id[:5], "")

	if err != nil {
		utils.Logger.Error("request CHECK", topo_id, err.Error())
		err_flag = 1
	}

	k8s_nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err1 != nil {
		utils.Logger.Error("can't check k8s cluster ", "corev1", err1.Error())
		err_flag = 1
	}

	for _, s := range k8s_nodes.Items {
		var hnode database.HostNode
		var hrm pb_common.InternalHostInfo

		h, err := database.FindHostEntity(topo_id[:5]+":"+s.Name, "")
		hnode = h
		if err != nil {
			utils.Logger.Error("request CHECK", topo_id[:5]+":"+s.Name, err.Error())
			err_flag = 1
		} else {
			hrm.Ip = hnode.Ip
			if hnode.Status == database.STATUS_READY {
				hrm.Status = pb_common.Status_READY

			} else {
				hrm.Status = pb_common.Status_NONE
			}

			hrm.RoutingRules = hnode.Routing_rule

			returnMessage.Hosts = append(returnMessage.Hosts, &hrm)
		}

	}

	for _, node := range topo.Vnodes {
		if strings.Contains(node.Name, "vhost") {
			var cnode database.ComputeNode
			var crm pb_common.InternalComputeInfo

			c, err := database.FindComputeEntity(topo_id[:5]+":"+node.Name, "")
			cnode = c
			if err != nil {
				utils.Logger.Error("request CHECK", topo_id[:5]+":"+node.Name, err.Error())
				err_flag = 1
			}

			crm.Id = cnode.Id
			crm.Name = cnode.Name
			crm.DatapathIp = cnode.DatapathIp
			crm.ContainerIp = cnode.ContainerIp
			crm.Mac = cnode.Mac
			crm.Veth = cnode.Veth
			if cnode.Status == database.STATUS_READY {
				crm.Status = pb_common.Status_READY
			} else if cnode.Status == database.STATUS_DELETING {
				crm.Status = pb_common.Status_DELETING
			} else if cnode.Status == database.STATUS_DEPLOYING {
				crm.Status = pb_common.Status_DEPLOYING
			} else {
				crm.Status = pb_common.Status_NONE
			}
			if cnode.OperationType == database.OPERATION_CREATE {
				crm.OperationType = pb_common.OperationType_CREATE
			} else if cnode.OperationType == database.OPERATION_INFO {
				crm.OperationType = pb_common.OperationType_INFO
			} else if cnode.OperationType == database.OPERATION_DELETE {
				crm.OperationType = pb_common.OperationType_DELETE
			} else if cnode.OperationType == database.OPERATION_UPDATE {
				crm.OperationType = pb_common.OperationType_UPDATE
			}

			returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &crm)

		}

	}

	go UpdateComputenodeInfo(k8client, topo_id, namespace)

	// err0 := UpdateComputenodeInfo(k8client, topo_id, namespace)

	// if err0 != nil {
	// 	utils.Logger.Error("can't update compute node info", topo_id, err0.Error())
	// 	err_flag = 1
	// }

	if err_flag == 1 {
		err_return = errs
		err_flag = 0
	} else {
		err_return = nil
	}
	return err_return
}

func Delete(k8client *kubernetes.Clientset, topo_id string, returnMessage *pb.ReturnTopologyMessage) error {
	topo, err_db := database.FindTopoEntity(topo_id[:5], "")

	if err_db != nil {
		utils.Logger.Error("request DELETE", topo_id, err_db.Error())
		return err_db
	}

	k8s_nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err1 != nil {

		utils.Logger.Error("request DELETE", "can't get k8s cluster node info", err1.Error())
		return err1
	}

	for _, s := range k8s_nodes.Items {
		var hnode database.HostNode
		var hrm pb_common.InternalHostInfo

		h, err := database.FindHostEntity(topo_id[:5]+":"+s.Name, "")

		hnode = h

		if err != nil {
			utils.Logger.Error("can't get host node info from DB error", topo_id[:5]+":"+s.Name, err.Error())
			return err
		} else {
			hrm.Ip = hnode.Ip
			if hnode.Status == database.STATUS_READY {
				hrm.Status = pb_common.Status_READY

			} else {
				hrm.Status = pb_common.Status_NONE
			}

			hrm.RoutingRules = hnode.Routing_rule

			returnMessage.Hosts = append(returnMessage.Hosts, &hrm)
		}

	}

	for _, node := range topo.Vnodes {
		if strings.Contains(node.Name, "vhost") {
			var cnode database.ComputeNode
			var crm pb_common.InternalComputeInfo

			c, err := database.FindComputeEntity(topo_id[:5]+":"+node.Name, "")

			cnode = c

			if err != nil {
				utils.Logger.Error("request DELETE", topo_id[:5]+":"+node.Name, err.Error())
			}

			crm.Id = cnode.Id
			crm.Name = cnode.Name
			crm.DatapathIp = cnode.DatapathIp
			crm.ContainerIp = cnode.ContainerIp
			crm.Mac = cnode.Mac
			crm.Veth = cnode.Veth
			crm.Status = pb_common.Status_DELETING
			crm.OperationType = pb_common.OperationType_DELETE

			returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &crm)

		}

	}

	namespace := "merak-" + topo_id[:5]

	go Topo_delete(k8client, topo, namespace)

	return nil
}

/*comment this function for query mac address*/
// func QueryMac(k8client *kubernetes.Clientset, topo_id string) error {

// 	topo_data, err := database.FindTopoEntity(topo_id[:5], "")

// 	if err != nil {
// 		return fmt.Errorf("failed to retrieve topology data from DB %s", err)
// 	}
// 	count_ping := 0

// 	for _, node := range topo_data.Vnodes {

// 		if strings.Contains(node.Name, "cgw") && count_ping < 1 {
// 			count_ping = count_ping + 1

// 			pod, err := database.FindPodEntity(topo_id+":"+node.Name, "")

// 			if err != nil {
// 				return fmt.Errorf("failed to retrieve pod data from DB %s", err)
// 			}

// 			ip_list, err_db := database.FindIPEntity(topo_id+":ips", "")
// 			if err_db != nil {
// 				return fmt.Errorf("failed to retrieve ip address data from DB %s", err_db)
// 			}

// 			for index := len(ip_list) - 1; index >= 0; index-- {
// 				ip := ip_list[index]
// 				ip = strings.Split(ip, "/")[0]
// 				cmd1 := []string{
// 					"ping",
// 					"-c 1",
// 					ip,
// 				}
// 				Pod_query(k8client, pod, cmd1)

// 			}

// 			cmd2 := []string{
// 				"arp",
// 				"-a",
// 			}
// 			out2, err3 := Pod_query(k8client, pod, cmd2)

// 			if err3 != nil {
// 				return fmt.Errorf("failed to query pod compute node info from K8s %s", err)
// 			}

// 			re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

// 			re2, _ := regexp.Compile(`^[a-fA-F0-9]{2}(:[a-fA-F0-9]{2}){5}$`)
// 			var ip string
// 			var mac string
// 			count := 0

// 			for _, s := range strings.Split(out2, " ") {
// 				s1 := strings.Trim(s, "(")
// 				s2 := strings.Trim(s1, ")")
// 				s2 = strings.Trim(s2, "\"")

// 				if re.MatchString(s2) {
// 					ip = s2
// 					count = count + 1
// 				} else if re2.MatchString(s) {
// 					mac = s2
// 					count = count + 1
// 					if count == 2 {
// 						database.SetValue(topo_id+":"+ip, mac)
// 						count = 0
// 					}
// 				}
// 			}
// 			if count != 0 {
// 				log.Printf("arp output parsing error. ip and mac doesn't match")
// 			}
// 		}
// 	}
// 	return nil
// }
