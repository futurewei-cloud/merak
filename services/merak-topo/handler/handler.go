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
	"strings"
	"time"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"

	pb_common "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//function CREATE
/* save the part of gw creation and mac learning for future requirment, comment the related code now*/
func Create(k8client *kubernetes.Clientset, topo_id string, aca_num uint32, rack_num uint32, aca_per_rack uint32, cgw_num uint32, data_plane_cidr string, ports_per_vswitch uint32, images []*pb.InternalTopologyImage, aca_parameters string, returnMessage *pb.ReturnTopologyMessage, topoPrefix string, namespace string) error {

	start_time := time.Now()

	var aca_image string
	var ovs_image string
	var err_return error

	for _, img := range images {
		if strings.Contains(img.Name, "ACA") {
			aca_image = img.Registry
		} else if strings.Contains(img.Name, "OVS") {
			ovs_image = img.Registry
		}
	}

	utils.Logger.Debug("request DEPLOY details", "Vhost number", aca_num, "Rack number", rack_num, "Vhosts per rack", aca_per_rack, "Ports per vswitch", ports_per_vswitch)

	topo, err_create := Create_multiple_layers_vswitches(int(aca_num), int(rack_num), int(aca_per_rack), int(ports_per_vswitch), data_plane_cidr)
	if err_create != nil {
		utils.Logger.Error("request DEPLOY", "multiple layers vswitches", err_create.Error())
		returnMessage.ReturnMessage = "Can not create multiple layers vswitches"
		returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
		return err_create

	}

	topo.Topology_id = topo_id

	elaps0 := time.Since(start_time)
	start0 := time.Now()

	utils.Logger.Info("request DEPLOY", " Complete: Generate topology data (in second)", elaps0)

	err_db := database.SetValue(topoPrefix, topo)
	if err_db != nil {
		utils.Logger.Error("request DEPLOY", "save topology to redis", err_db.Error(), "topo_id", topoPrefix)
		returnMessage.ReturnMessage = "Can not save topology to redis"
		returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
		return err_db
	}

	elaps1 := time.Since(start0)
	start1 := time.Now()
	utils.Logger.Info("request DEPLOY", "Save topology in redis DB (in second)", elaps1)

	k8s_nodes, err2 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err2 != nil {
		utils.Logger.Error("request DEPLOY", "k8s cluster nodes no response", err2.Error())
		returnMessage.ReturnMessage = "CREATE: can not list k8s cluster worker nodes info"
		returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
		return err2
	}

	for _, s := range k8s_nodes.Items {
		var hnode database.HostNode

		node_yaml, err3 := k8client.CoreV1().Nodes().Get(Ctx, s.Name, metav1.GetOptions{})
		if err3 != nil {
			utils.Logger.Error("can't get host node info in k8s cluster", s.Name, err3.Error())
			returnMessage.ReturnMessage = "DEPLOY: can not get k8s host node info from k8s cluster"
			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			return err3
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
		err := database.SetValue(topoPrefix+":"+s.Name, hnode)
		if err != nil {
			utils.Logger.Error("can not save host node info in DB", "key", topoPrefix+":"+s.Name, "error msg", err.Error())
			returnMessage.ReturnMessage = "DEPLOY: can not save host node info in DB"
			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
			return err
		}

	}

	for _, s := range k8s_nodes.Items {
		var hnode database.HostNode
		var hrm pb_common.InternalHostInfo

		h, err := database.FindHostEntity(topoPrefix+":"+s.Name, "")

		hnode = h

		if err != nil {
			utils.Logger.Error("cannot query host node info from DB ", topoPrefix+":"+s.Name, err.Error())
			returnMessage.ReturnMessage = "DEPLOY: cannot query host node info from DB"
			returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
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

			err := database.SetValue(topoPrefix+":"+cnode.Name, cnode)

			if err != nil {
				utils.Logger.Error("can not save compute node info in DB ", topoPrefix+":"+cnode.Name, err.Error())
				returnMessage.ReturnMessage = "DEPLOY: can not save compute node info in DB"
				returnMessage.ReturnCode = pb_common.ReturnCode_FAILED
				return err
			}

			crm.Id = cnode.Id
			crm.Name = cnode.Name
			crm.DatapathIp = cnode.DatapathIp
			crm.ContainerIp = cnode.ContainerIp
			crm.Mac = cnode.Mac
			crm.Veth = cnode.Veth
			crm.Status = pb_common.Status_DEPLOYING
			crm.OperationType = pb_common.OperationType_CREATE

			returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &crm)

		}

	}

	for _, img := range images {
		if strings.Contains(img.Name, "ACA") {
			aca_image = img.Registry
		} else if strings.Contains(img.Name, "OVS") {
			ovs_image = img.Registry
		}
	}

	if namespace != "default" {
		nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
		k8client.CoreV1().Namespaces().Create(context.Background(), nsSpec, metav1.CreateOptions{})
		utils.Logger.Info("request DEPLOY", "create k8s cluster namespace for new topology deployment", namespace)
	}

	go Topo_deploy(k8client, aca_image, ovs_image, topo, aca_parameters, topoPrefix, namespace)

	elaps2 := time.Since(start1)

	utils.Logger.Info("request DEPLOY", "Complete: Return compute nodes and k8s cluster information to Scenario Manager (in second)", elaps2)
	return nil

}

func UpdateComputenodeInfo(k8client *kubernetes.Clientset, topoPrefix string, namespace string) error {

	start_time := time.Now()
	topo, err := database.FindTopoEntity(topoPrefix, "")

	if err != nil {
		utils.Logger.Warn("can't find topology info in DB", topoPrefix, err.Error())
		return err
	}

	k8s_nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err1 != nil {
		utils.Logger.Error("can't check k8s cluster ", "corev1", err1.Error())
		return err1
	}

	for _, s := range k8s_nodes.Items {
		var hnode database.HostNode

		node_yaml, err2 := k8client.CoreV1().Nodes().Get(Ctx, s.Name, metav1.GetOptions{})
		if err2 != nil {
			utils.Logger.Error("can't get k8s host node info ", s.Name, err2.Error())
			return err2
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
		err := database.SetValue(topoPrefix+":"+s.Name, hnode)
		if err != nil {
			utils.Logger.Warn("can't save host node in DB", topoPrefix+":"+s.Name, err.Error())
			return err
		}

	}

	elaps0 := time.Since(start_time)
	start_time0 := time.Now()

	utils.Logger.Info("Complete", "updating host node info in DB (in second) ", elaps0)

	for _, node := range topo.Vnodes {

		if strings.Contains(node.Name, "vhost") {

			var cnode database.ComputeNode

			res, err := k8client.CoreV1().Pods(namespace).Get(Ctx, node.Name, metav1.GetOptions{})

			if err != nil {
				utils.Logger.Error("can't get pod info from k8s", node.Name, err.Error(), "namespace", namespace)
				return err
			} else {
				cnode.Name = res.Name
				cnode.Id = string(res.UID)

				cnode.DatapathIp = strings.Split(node.Nics[len(node.Nics)-1].Ip, "/")[0]
				cnode.Veth = node.Nics[len(node.Nics)-1].Intf
				if res.Status.PodIP != "" {
					cnode.ContainerIp = res.Status.PodIP
					cnode.HostName = res.Spec.NodeName
				} else {
					utils.Logger.Debug("Warning", "pod ip is not ready", res.Name)
				}

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

				err_db := database.SetValue(topoPrefix+":"+node.Name, cnode)
				if err_db != nil {
					utils.Logger.Warn("can't save compute node in DB", topoPrefix+":"+node.Name, err_db.Error())
					return err_db
				}

			}

		}

	}

	elaps := time.Since(start_time0)

	utils.Logger.Info("Complete", "updating compute nodes info in DB (in second) ", elaps)

	return nil
}

func Info(k8client *kubernetes.Clientset, topo_id string, returnMessage *pb.ReturnTopologyMessage, topoPrefix string, namespace string) error {

	topo, err := database.FindTopoEntity(topoPrefix, "")

	if err != nil {
		utils.Logger.Warn("request CHECK", topoPrefix, err.Error())
		return err
	}

	k8s_nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err1 != nil {
		utils.Logger.Error("k8s cluster no response ", "corev1", err1.Error())
		return err1
	}

	for _, s := range k8s_nodes.Items {
		var hnode database.HostNode
		var hrm pb_common.InternalHostInfo

		hnode, err := database.FindHostEntity(topoPrefix+":"+s.Name, "")

		if err != nil {
			utils.Logger.Warn("request CHECK", topoPrefix+":"+s.Name, err.Error())
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

			cnode, err := database.FindComputeEntity(topoPrefix+":"+node.Name, "")

			if err != nil {
				utils.Logger.Warn("request CHECK", topoPrefix+":"+node.Name, err.Error())
				return err
			}
			crm.Id = cnode.Id
			crm.Name = cnode.Name
			crm.DatapathIp = cnode.DatapathIp
			crm.ContainerIp = cnode.ContainerIp
			crm.Mac = cnode.Mac
			crm.Veth = cnode.Veth
			crm.Hostname = cnode.HostName

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

	go UpdateComputenodeInfo(k8client, topoPrefix, namespace)

	return nil
}

func Delete(k8client *kubernetes.Clientset, topo_id string, returnMessage *pb.ReturnTopologyMessage, topoPrefix string, namespace string) error {

	topo, err_db := database.FindTopoEntity(topoPrefix, "")

	if err_db != nil {
		utils.Logger.Warn("request DELETE", "can't query topology data from DB", topoPrefix, err_db.Error())
		return err_db
	}

	k8s_nodes, err1 := k8client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err1 != nil {
		utils.Logger.Error("request DELETE", "k8s cluster no response", err1.Error())
		return err1
	}

	for _, s := range k8s_nodes.Items {
		var hnode database.HostNode
		var hrm pb_common.InternalHostInfo

		h, err := database.FindHostEntity(topoPrefix+":"+s.Name, "")

		hnode = h

		if err != nil {
			utils.Logger.Error("can't get host node info from DB error", topoPrefix+":"+s.Name, err.Error())
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

			cnode, err := database.FindComputeEntity(topoPrefix+":"+node.Name, "")

			if err != nil {
				utils.Logger.Error("request DELETE", topoPrefix+":"+node.Name, err.Error())
				return err
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

	go Topo_delete(k8client, topo, topoPrefix, namespace)

	return nil
}

