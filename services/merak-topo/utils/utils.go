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
	"github.com/futurewei-cloud/merak/services/common/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	Logger logger.Logger
)

const K8SAPIURL string = "https://kubernetes.default:6443"

func Init_logger() {
	var err error

	Logger, err = logger.NewLogger(logger.DEBUG, "Merak-topo")

	if err != nil {
		Logger.Fatal("Can not build a new logger", err)
	} else {
		Logger.Info("Create logger for merak-topo", "Init_logger", "Created")
	}

}

func K8sClient() (*kubernetes.Clientset, error) {

	config, err_config := K8sConfig()
	if err_config != nil {
		Logger.Error("can't create k8s client", "k8s client initiate error", err_config.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Logger.Fatal("can't create k8s client with config", "k8s clientset with config initiate error", err.Error())
	}

	return clientset, err

}

func K8sConfig() (*rest.Config, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		Logger.Fatal("can't create k8s config", "k8s config initiate error", err.Error())
	}

	return config, nil

}
