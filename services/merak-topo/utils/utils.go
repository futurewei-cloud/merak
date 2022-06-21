package utils

import (
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/util/homedir"

	clientcmd "k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/kubernetes"
)

func K8sClient() (*kubernetes.Clientset, error) {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return clientset, err

}
