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

type ServiceStatus string

const (
	STATUS_NONE      ServiceStatus = "NONE"
	STATUS_DEPLOYING ServiceStatus = "DEPLOYING"
	STATUS_READY     ServiceStatus = "READY"
	STATUS_DELETING  ServiceStatus = "DELETING"
	STATUS_UPDATING  ServiceStatus = "UPDATING"
	STATUS_FAILED    ServiceStatus = "FAILED"
	STATUS_DONE      ServiceStatus = "DONE"
)

type Nic struct {
	Id   string `json:"id"`
	Intf string `json:"intf"`
	Ip   string `json:"ip"`
	Mac  string `json:"mac"`
}

type Vnode struct {
	Id          string        `json:"id"`
	Type        string        `json:"type"`
	Name        string        `json:"name"`
	Nics        []Nic         `json:"nics"`
	Flinks      []Vlink       `json:"flinks"`
	ContainerIp string        `json:"containerip"`
	Status      ServiceStatus `json:"status"`
}

type Vlink struct {
	Id         string        `json:"id"`
	Name       string        `json:"name"`
	Uid        int           `json:"uid"`
	Peer_pod   string        `json:"peer_pod"`
	Local_pod  string        `json:"local_pod"`
	Local_intf string        `json:"local_intf"`
	Local_ip   string        `json:"local_ip"`
	Peer_intf  string        `json:"peer_intf"`
	Peer_ip    string        `json:"peer_ip"`
	Status     ServiceStatus `json:"status"`
}

type TopologyData struct {
	Topology_id string  `json:"topology_id"`
	Vnodes      []Vnode `json:"vnodes"`
}
