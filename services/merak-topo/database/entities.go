package database

type Nic struct {
	Intf string `json:"intf"`
	Ip   string `json:"ip"`
}

type Vnode struct {
	Name string `json:"name"`
	Nics []Nic  `json:"nics"`
}

type Vport struct {
	Name string `json:"name"`
	Intf string `json:"nic"`
	Ip   string `json:"ip"`
}

type Vlink struct {
	Name string `json:"name"`
	Src  Vport  `json:"src"`
	Dst  Vport  `json:"dst"`
}

type TopologyData struct {
	Topology_id string  `json:"topology_id"`
	Vnodes      []Vnode `json:"vnodes"`
	Vlinks      []Vlink `json:"vlinks"`
}
