package handler

type ConfigClink struct {
	Local_pod  string
	Local_intf string
	Local_ip   string
	Peer_intf  string
	Peer_pod   string
	Peer_ip    string
}

type Clink struct {
	Uid        string
	Local_intf string
	Local_ip   string
	Peer_intf  string
	Peer_pod   string
	Peer_ip    string
}

type LabelStruct struct {
	Topo string
}

type Cmetadata struct {
	Name   string
	labels LabelStruct
}

type Clinks struct {
	Links []Clink
}

type Cnode struct {
	ApiVersion string
	Kind       string
	Metadata   Cmetadata
	Spec       Clinks
}
