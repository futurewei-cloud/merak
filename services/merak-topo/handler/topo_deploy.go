package handler

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func Topo_deploy() {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})

	config, err := kubeconfig.ClientConfig()

	if err != nil {
		panic(err)
	}

	clientset := kubernetes.NewForConfigOrDie(config)

	modeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})

	if err != nil {
		panic(err)
	}
	for _, n := range modeList.Items {
		fmt.Println(n.Name)
	}

	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test1box", Image: "busybox:latest", Command: []string{"sleep", "100000"}},
			},
		},
	}

	pod, err := clientset.CoreV1().Pods("default").Create(context.Background(), newPod, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
	fmt.Println(pod)

}
