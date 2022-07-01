package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	// logrus "github.com/sirupsen/logrus"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/remotecommand"
)

var (
	// ACA_IMAGE = "phudtran/aca:latest"
	ACA_IMAGE = "phudtran/merak-agent:dev"
	OVS_IMAGE = "yanmo96/ovs_only:latest"
	GW_IMAGE  = "yanmo96/aca_build_standard:v2"
	Ctx       = context.Background()

	namespace        = "default"
	topologyClassGVR = schema.GroupVersionResource{
		Group:    "networkop.co.uk",
		Version:  "v1beta1",
		Resource: "topologies",
	}
)

func CreateTopologyClasses(client dynamic.Interface, name string, links []map[string]interface{}) error {
	rc := NewTopologyClass(name, links)

	_, err := client.Resource(topologyClassGVR).Namespace(namespace).Create(Ctx, rc, metav1.CreateOptions{})

	log.Printf("Creating TopologyClass %s", name)

	if err != nil {
		return fmt.Errorf("failed to create topologyClass %s", err)
	}

	return nil

}

func GetTopologyClasses(client dynamic.Interface, name string) error {

	data, err := client.Resource(topologyClassGVR).Namespace(namespace).Get(Ctx, name, metav1.GetOptions{})

	log.Printf("Get TopologyClass %s", name)

	if err != nil {
		return fmt.Errorf("failed to create topologyClass %s", err)
	}

	fmt.Printf("Get %v topology data: %v", name, data)

	return nil

}

func DeleteTopologyClasses(client dynamic.Interface, name string) error {

	err := client.Resource(topologyClassGVR).Namespace(namespace).Delete(Ctx, name, metav1.DeleteOptions{})

	log.Printf("Delete TopologyClass %s", name)

	if err != nil {
		return fmt.Errorf("failed to create topologyClass %s", err)
	}

	return nil

}

func NewTopologyClass(name string, links []map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "Topology",
			"apiVersion": "networkop.co.uk/v1beta1",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"links": links,
			},
		},
	}
}

func Topo_deploy(k8client *kubernetes.Clientset, topo database.TopologyData) error {

	nodes := topo.Vnodes

	// dynamic client

	config := ctrl.GetConfigOrDie()
	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client %s", err)
	}

	for _, node := range nodes {

		// Create topology class

		err := CreateTopologyClasses(dclient, node.Name, node.Flinks)

		if err != nil {
			return fmt.Errorf("failed to create runtime class %s", err)
		}

		//// create pods
		var newPod *corev1.Pod
		if strings.Contains(node.Name, "vhost") {
			newPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: node.Name,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "vhost", Image: ACA_IMAGE, Command: []string{"sleep", "100000"}},
					},
				},
			}
		} else if strings.Contains(node.Name, "vswitch") || strings.Contains(node.Name, "tor") {
			newPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: node.Name,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "vswitch", Image: OVS_IMAGE, Command: []string{"sleep", "100000"}},
					},
				},
			}
		} else if strings.Contains(node.Name, "cgw") {
			newPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: node.Name,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "cgw", Image: GW_IMAGE, Command: []string{"sleep", "100000"}},
					},
				},
			}

		} else {
			return errors.New("no image for this device, please upload the image before create topology")
		}

		_, err = k8client.CoreV1().Pods("default").Create(Ctx, newPod, metav1.CreateOptions{})

		if err != nil {
			return fmt.Errorf("create pod error %s", err)
		} else {
			err = database.SetValue(node.Name+"-pod", newPod)
			if err != nil {
				log.Fatalf("fail to save topology in DB %s", err)
			}

		}

		//save yaml file during deployment processes

	}

	// save comput node data

	// var comput_nodes *pb.InternalComputeInfo

	return nil

}

func Pod_info(k8client *kubernetes.Clientset, pod *corev1.Pod) error {

	ip := "10.200.99.11"
	command := "arp " + ip
	// command2 := "arp " + ip
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := k8client.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).Namespace(pod.ObjectMeta.Namespace).SubResource("exec") // .Param("container", containerName)
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		panic(err.Error())
	}
	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&corev1.PodExecOptions{
		Stdin:  false,
		Stdout: true,
		Stderr: true,
		TTY:    false,
		// Container: containerName,
		Container: pod.Name,
		Command:   cmd,
	}, parameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(&rest.Config{}, "POST", req.URL())
	if err != nil {
		panic(err.Error())
	}
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		panic(err.Error())
	}
	log.Printf("Output from pod: %v", stdout.String())
	log.Printf("Error from pod: %v", stderr.String())
	return nil

}

func Topo_delete(k8client *kubernetes.Clientset, topo_id string) error {

	topo_data, err := database.FindTopoEntity(topo_id, "")

	if err != nil {
		return fmt.Errorf("failed to get topology data from DB %s", err)
	}

	config := ctrl.GetConfigOrDie()
	dclient, err := dynamic.NewForConfig(config)

	if err != nil {
		return fmt.Errorf("failed to create dynamic client %s", err)
	}

	for _, node := range topo_data.Vnodes {

		err := k8client.CoreV1().Pods("default").Delete(Ctx, node.Name, metav1.DeleteOptions{})

		if err != nil {
			return fmt.Errorf("delete pod container error %s", err)
		}

		err = DeleteTopologyClasses(dclient, node.Name)
		if err != nil {
			return fmt.Errorf("delete pod topology error %s", err)
		}
	}
	return nil
}

// save topology to redis
func Topo_save(k8client *kubernetes.Clientset, topo database.TopologyData) error {
	// check pod status
	topo_id := topo.Topology_id
	pods_status, err := Topo_pod_check(k8client, topo)
	if err != nil {
		return fmt.Errorf("check topology pod error %s", err)
	}

	if pods_status {
		fmt.Println("All pods are created")
		database.SetValue(topo_id, topo)
	} else {
		fmt.Println("Pods are pending")
	}

	return err
}

func Topo_pod_check(k8client *kubernetes.Clientset, topo database.TopologyData) (bool, error) {

	s := true

	nodes := topo.Vnodes

	for _, node := range nodes {

		res, err := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})

		if err != nil {
			return false, fmt.Errorf("get pod error %s", err)
		}
		if res.Status.Phase != "Running" {
			s = false
		}
		fmt.Printf("Get information of node %v: %v", node.Name, res.Status)

	}
	return s, nil
}

func Comput_node_info(k8client *kubernetes.Clientset, topo database.TopologyData) ([]*pb.InternalComputeInfo, error) {

	var cnodes []*pb.InternalComputeInfo
	var cnode *pb.InternalComputeInfo

	for _, node := range topo.Vnodes {

		res, err := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})

		if err != nil {
			return cnodes, fmt.Errorf("get pod error %s", err)
		}
		if res.Status.Phase == "Running" {

			// out, _ := k8client.CoreV1().Pods("default").Get(Ctx, node.Name, metav1.GetOptions{})
			out, err1 := k8client.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
			if err1 != nil {
				return cnodes, fmt.Errorf("get pod data error %s", err)
			}

			for i := range out.Items {

				cnode.Name = out.Items[i].Status.NominatedNodeName
				cnode.Id = out.Items[i].Status.HostIP
				// cnode.HostIP = out.Items[i].Status.HostIP
				cnode.Ip = out.Items[i].Status.PodIP
				cnode.Mac = ""
				cnode.Veth = ""
				cnode.OperationType = 2
				cnodes = append(cnodes, cnode)
			}
		}
	}
	return cnodes, nil
}
