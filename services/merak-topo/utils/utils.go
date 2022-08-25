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

package utils

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// const K8SAPIURL string = "https://kubernetes.default:6443"
// const K8SAPIURL string = "https://172.31.28.160:6443"
const K8SAPIURL string = "https://172.31.28.160:6443"

const K8SCONFIGPATH string = "/etc/kubernetes/admin.conf"

func K8sClient() (*kubernetes.Clientset, error) {

	config, err_config := K8sConfig()
	if err_config != nil {
		return nil, fmt.Errorf(err_config.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return clientset, err

}

// func K8sConfig() (*rest.Config, error) {

// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	return config, nil

// }

func K8sConfig() (*rest.Config, error) {

	config, err := clientcmd.BuildConfigFromFlags(K8SAPIURL, K8SCONFIGPATH)
	if err != nil {
		return nil, err
	}

	return config, nil

}
