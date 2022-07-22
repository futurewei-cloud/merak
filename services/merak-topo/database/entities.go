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
	Intf string `json:"intf"`
	Ip   string `json:"ip"`
	Mac  string `json:"mac"`
}

type Vnode struct {
	Id          string                   `json:"id"`
	Name        string                   `json:"name"`
	Nics        []Nic                    `json:"nics"`
	Flinks      []map[string]interface{} `json:"flinks"`
	ContainerIp string                   `json:"containerip"`
	Status      ServiceStatus            `json:"status"`
}

type Vport struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Intf string `json:"nic"`
	Ip   string `json:"ip"`
}

type Vlink struct {
	Id     string        `json:"id"`
	Name   string        `json:"name"`
	Src    Vport         `json:"src"`
	Dst    Vport         `json:"dst"`
	Status ServiceStatus `json:"status"`
}

type TopologyData struct {
	Topology_id string  `json:"topology_id"`
	Vnodes      []Vnode `json:"vnodes"`
	Vlinks      []Vlink `json:"vlinks"`
}
