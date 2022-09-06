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
	"fmt"

	"testing"

	"github.com/futurewei-cloud/merak/services/merak-topo/database"
)

func TestDB(t *testing.T) {
	database.ConnectDatabase()
	fmt.Println("DB created.")

	// nic1 := database.Nic{
	// 	Intf: "a1-intf1",
	// 	Ip:   "10.99.1.2",
	// }
	// nic2 := database.Nic{
	// 	Intf: "a1-intf2",
	// 	Ip:   "10.99.1.3",
	// }

	// nic3 := database.Nic{
	// 	Intf: "a2-intf1",
	// 	Ip:   "10.99.1.4",
	// }

	// vnode0 := database.Vnode{
	// 	Name: "node-id1",
	// 	Nics: []database.Nic{nic1, nic2},
	// }

	// vnode1 := database.Vnode{
	// 	Name: "node-id2",
	// 	Nics: []database.Nic{nic3},
	// }

	// local1 := database.Vport{
	// 	Name: "port-id1",
	// 	Intf: vnode1.Nics[len(vnode1.Nics)-1].Intf,
	// 	Ip:   vnode1.Nics[len(vnode1.Nics)-1].Ip,
	// }
	// peer1 := database.Vport{
	// 	Name: "port-id2",
	// 	Intf: vnode0.Nics[len(vnode0.Nics)-1].Intf,
	// 	Ip:   vnode0.Nics[len(vnode0.Nics)-1].Ip,
	// }

	// vlink1 := database.Vlink{
	// 	Name: "link-id1",
	// 	Src:  local1,
	// 	Dst:  peer1,
	// }

	// topo1 := database.TopologyData{
	// 	Topology_id: "topo1",
	// 	Vnodes:      []database.Vnode{vnode0, vnode1},
	// 	Vlinks:      []database.Vlink{vlink1},
	// }

	// fmt.Println("set topo1")
	// database.SetValue("topo:proj1-topo1", topo1)
	// fmt.Println("set node1")
	// database.SetValue("node:proj1-topo1-node1", vnode0)
	// fmt.Println("set link1")
	// database.SetValue("link:proj1-topo1-link1", vlink1)

	val, _ := database.Get("topo:proj1-topo1")

	fmt.Printf("The returned data from Radis is %+v. \n", val)

	val, _ = database.Get("node:proj1-topo1-node1")

	fmt.Printf("The returned data from Radis is %+v.\n ", val)

	val, _ = database.Get("link:proj1-topo1-link1")

	fmt.Printf("The returned data from Radis is %+v.\n", val)

}
