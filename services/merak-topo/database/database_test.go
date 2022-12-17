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

package database

import (
	"fmt"
	"testing"

	entities "github.com/futurewei-cloud/merak/services/merak-topo/entities"
)

func TestDBConnection(t *testing.T) {
	ConnectDatabase()
	fmt.Println("DB created.")

}

func TestSetValue_ComputeNode(t *testing.T) {
	var vhost_test entities.ComputeNode
	vhost_test.ContainerIp = "10.100.0.1"
	vhost_test.DatapathIp = "10.200.0.0/16"
	vhost_test.Id = "vhost_test_id"
	vhost_test.Name = "vhost_test"
	vhost_test.Mac = "0000001"
	vhost_test.Veth = "eth-1"
	vhost_test.OperationType = "CREATE"
	vhost_test.Status = "DEPLOYING"

	err := SetValue("topo_test_vhost", vhost_test)
	if err != nil {
		fmt.Println("SetValue for ComputeNode passed")
	} else {
		fmt.Println("SetValue for ComputeNode failed")
	}
}

func TestSetValue_HostNode(t *testing.T) {
	var host_test entities.HostNode
	host_test.Routing_rule[0] = "default via 10.0.0.1 dev wlp3s0 proto static metric 600"
	host_test.Routing_rule[1] = "10.0.0.0/24 dev wlp3s0 proto kernel scope link src 10.0.0.73 metric 600"
	host_test.Ip = "10.0.0.2"
	host_test.Status = "READY"

	err := SetValue("topo_test_host", host_test)
	if err != nil {
		fmt.Println("SetValue for ComputeNode passed")
	} else {
		fmt.Println("SetValue for ComputeNode failed")
	}
}
