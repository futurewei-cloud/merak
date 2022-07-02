package handler

import (
	"github.com/futurewei-cloud/merak/services/merak-topo/database"

	"fmt"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// grpc client

// Call operation.Create()
// --- Generete topology yaml file based on topology type

// Call operation.SaveTopo()
// ----Save topology yaml file

// Call operation.ConfigTopo()
// ---generate pod topology
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
func Create(k8client *kubernetes.Clientset, topo_id string, aca_num uint32, rack_num uint32, aca_per_rack uint32, data_plane_cidr string, returnMessage *pb.ReturnTopologyMessage) error {

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

	fmt.Println("======== Pairing links ==== ")

	Links_gen(Topo_nodes)
	// fmt.Printf("The topology links are : %+v. \n", Topo_links)

	fmt.Println("======== Generate topology data ==== ")
	Topo.Topology_id = topo_id
	Topo.Vlinks = Topo_links
	Topo.Vnodes = Topo_nodes

	fmt.Println("======== Topology Deployment ==== ")

	err := Topo_deploy(k8client, Topo)
	if err != nil {
		return fmt.Errorf("topology deployment error %s", err)
	}

	fmt.Println("======== Save topo to redis =====")
	err1 := Topo_save(k8client, Topo)
	if err1 != nil {
		return fmt.Errorf("save topo to redis error %s", err1)
	}

	fmt.Println("========= Get compute nodes information after deployment=====")

	for _, node := range Topo.Vnodes {
		var cnode pb.InternalComputeInfo

		res, err := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})

		if err != nil {
			return fmt.Errorf("get pod error %s", err)
		} else {
			// if res.Status.Phase == "Running" && res.Labels["Type"] == "vhost" {
			if res.Labels["Type"] == "vhost" {
				cnode.Name = res.Name
				cnode.Id = string(res.UID)
				// cnode.HostIP = res.Status.HostIP
				cnode.Ip = res.Status.PodIP
				cnode.Mac = ""
				cnode.Veth = ""
				cnode.OperationType = pb.OperationType_INFO
				returnMessage.ComputeNodes = append(returnMessage.ComputeNodes, &cnode)
			}
		}
	}
	return nil
}

func Info(k8client *kubernetes.Clientset, topo_id string, returnMessage *pb.ReturnTopologyMessage) error {

	topo, err := database.FindTopoEntity(topo_id, "")

	if err != nil {
		return fmt.Errorf("query topology_id error %s", err)
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
				// cnode.HostIP = res.Status.HostIP
				cnode.Ip = res.Status.PodIP
				cnode.Mac = ""
				cnode.Veth = ""
				cnode.OperationType = pb.OperationType_INFO
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

func Subtest(k8client *kubernetes.Clientset, topo_id string) error {

	topo_data, err := database.FindTopoEntity(topo_id, "")

	if err != nil {
		return fmt.Errorf("failed to retrieve topology data from DB %s", err)
	}

	for _, node := range topo_data.Vnodes {

		pod, err := database.FindPodEntity(node.Name, "-pod")

		if err != nil {
			return fmt.Errorf("failed to retrieve pod data from DB %s", err)
		}

		err = Pod_info(k8client, pod)
		if err != nil {
			return fmt.Errorf("failed to query pod compute node info from K8s %s", err)
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
