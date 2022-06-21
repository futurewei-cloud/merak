package database

type Nic struct {
	Intf string `json:"intf"`
	Ip   string `json:"ip"`
}

type Vnode struct {
	Id     string        `json:"id"`
	Name   string        `json:"name"`
	Nics   []Nic         `json:"nics"`
	Flinks []ConfigClink `json:"flinks"`
}

type Vport struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Intf string `json:"nic"`
	Ip   string `json:"ip"`
}

type Vlink struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Src  Vport  `json:"src"`
	Dst  Vport  `json:"dst"`
}

type TopologyData struct {
	Topology_id string  `json:"topology_id"`
	Vnodes      []Vnode `json:"vnodes"`
	Vlinks      []Vlink `json:"vlinks"`
}

type ConfigClink struct {
	Uid        string
	Local_intf string
	Local_ip   string
	Peer_pod   string
	Peer_intf  string
	Peer_ip    string
}

// cmap {
// 	'apiVersion': 'networkop.co.uk/v1beta1',
// 	'kind': 'Topology',
// 	'metadata': {
// 		'name': 'ovs-0',
// 		'labels': {'topo': 'demo_v2aca'}
// 		},
// 	'spec': {
// 		'links': [
// 			{
// 				'uid': 0,
// 				'local_intf': 'eth1',
// 				'local_ip': '',
// 				'peer_intf': 'eth1',
// 				'peer_pod': 'aca-0',
// 				'peer_ip': '10.99.0.1/24'
// 			},
// 			{
// 				'uid': 1,
// 				'local_intf': 'eth2',
// 				'local_ip': '',
// 				'peer_intf': 'eth1',
// 				'peer_pod': 'aca-1',
// 				'peer_ip': '10.99.0.2/24'
// 			},
// 			{
// 				'uid': 2,
// 				'local_intf': 'eth3',
// 				'local_ip': '',
// 				'peer_intf': 'eth1',
// 				'peer_pod': 'aca-2',
// 				'peer_ip': '10.99.0.3/24'
// 			},
// 			{
// 				'uid': 3,
// 				'local_intf': 'eth4',
// 				'local_ip': '',
// 				'peer_intf': 'eth1',
// 				'peer_pod': 'aca-3',
// 				'peer_ip': '10.99.0.4/24'
// 			}
// 		]}}
