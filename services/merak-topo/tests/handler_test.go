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

package tests

import (
	"context"
	"testing"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
	"github.com/futurewei-cloud/merak/services/merak-topo/handler"

	"github.com/futurewei-cloud/merak/services/merak-topo/utils"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type k8s struct {
	clientset kubernetes.Interface
}

func (o *k8s) getVersion() (string, error) {
	version, err := o.clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

func newTestSimpleK8s() *k8s {
	client := k8s{}
	client.clientset = fake.NewSimpleClientset()
	return &client
}

func TestK8sEnv(t *testing.T) {
	k8s := newTestSimpleK8s()
	_, err := k8s.getVersion()
	if err != nil {
		t.Fatal("k8s getVersion should not raise an error")
	}
	assert.Nil(t, err)
}

func TestTopoHandlerWorkFlowSuccess(t *testing.T) {
	aca_num := 12
	rack_num := 3
	aca_per_rack := 4
	data_plane_cidr := "10.200.0.0/16"

	ports_per_vswitch := 3
	topoPrefix := "1topo"
	utils.Init_logger()

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	database.Rdb = client
	topo, err1 := handler.Create_multiple_layers_vswitches(int(aca_num), int(rack_num), int(aca_per_rack), int(ports_per_vswitch), data_plane_cidr)
	err2 := database.SetValue(topoPrefix, topo)

	k8s := newTestSimpleK8s()
	_, err := k8s.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	assert.Nil(t, err1, err2, err)

}

func TestTopoHandlerWorkFlowFail(t *testing.T) {
	aca_num := 20
	rack_num := 25
	aca_per_rack := 4
	data_plane_cidr := "10.200.0.0/16"

	ports_per_vswitch := 3
	topoPrefix := "1topo"
	utils.Init_logger()

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	database.Rdb = client
	topo, err1 := handler.Create_multiple_layers_vswitches(int(aca_num), int(rack_num), int(aca_per_rack), int(ports_per_vswitch), data_plane_cidr)
	err2 := database.SetValue(topoPrefix, topo)

	k8s := newTestSimpleK8s()
	_, err := k8s.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	assert.Error(t, err1, err2, err)

}
