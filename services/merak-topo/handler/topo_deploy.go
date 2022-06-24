package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var (
	ACA_IMAGE = "phudtran/aca:latest"
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

	///////////// dynamic client

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

		///////// create pods
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
		}

	}

	// check pod status
	pods_status, err1 := Topo_pod_check(k8client, topo)
	if err1 != nil {
		return fmt.Errorf("check topology pod error %s", err1)
	}

	if pods_status {
		// save topology to redis
		fmt.Println("All pods are created")
	}

	return nil

}

func Topo_delete(k8client *kubernetes.Clientset, topo database.TopologyData) error {

	nodes := topo.Vnodes

	config := ctrl.GetConfigOrDie()
	dclient, err := dynamic.NewForConfig(config)

	if err != nil {
		return fmt.Errorf("failed to create dynamic client %s", err)
	}

	for _, node := range nodes {

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
