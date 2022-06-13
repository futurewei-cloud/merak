package entities

import (
	"time"
)

type ScenarioStatus string

const (
	STATUS_NONE      ScenarioStatus = "NONE"
	STATUS_DEPLOYING ScenarioStatus = "DEPLOYING"
	STATUS_READY     ScenarioStatus = "READY"
	STATUS_DELETING  ScenarioStatus = "DELETING"
)

type EventName string

const (
	ENENT_DEPLOY EventName = "DEPLOY"
	EVENT_DELETE EventName = "DELETE"
	EVENT_CHECK  EventName = "CHECK"
)

// Action
type Event struct {
	Id         string    `json:"id"`
	ScenarioId string    `json:"scenario_id"`
	Action     EventName `json:"action"`
}

// Scenario
type Scenario struct {
	Id            string         `json:"id"`
	Name          string         `json:"name"`
	ProjectId     string         `json:"project_id"`
	TopologyId    string         `json:"topology_id"`
	ServiceConfId string         `json:"service_config_id"`
	NetworkConfId string         `json:"network_config_id"`
	ComputeConfId string         `json:"compute_config_id"`
	TestConfId    string         `json:"test_config_id"`
	Status        ScenarioStatus `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Service Configuration
type ServiceConfig struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Services  []Service `json:"services"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Service struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Cmd          string   `json:"cmd"`
	Url          string   `json:"url"`
	Parameters   []string `json:"parameters"`
	ReturnCode   []uint32 `json:"return_code"`
	ReturnString []string `json:"return_string"`
	WhenToRun    string   `json:"when_to_run"`
	WhereToRun   string   `json:"where_to_run"`
}

// Topology Configuration
type TopologyConfig struct {
	Id                           string    `json:"id"`
	Name                         string    `json:"name"`
	TopoType                     string    `json:"type"`
	NumberOfVhosts               uint      `json:"number_of_vhosts"`
	NumberOfRacks                uint      `json:"number_of_racks"`
	VhostsPerRack                uint      `json:"vhosts_per_rack"`
	NumberOfControlPlaneGateways uint      `json:"number_of_control_plane_gateways"`
	ControlPlaneGatewayIPs       []string  `json:"control_plane_gateway_ips"`
	Images                       []Image   `json:"images"`
	VNodes                       []VNode   `json:"vnodes"`
	VLinks                       []VLink   `json:"vlinks"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

type Image struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Registry string   `json:"registry"`
	Cmd      []string `json:"cmd"`
	Args     []string `json:"args"`
}

type VNode struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Nics []Nic  `json:"nics"`
}

type Nic struct {
	Name string `json:"name"`
	Ip   string `json:"ip"`
}

type VLink struct {
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`
}

// Network Configuration
type NetworkConfig struct {
	Id                     string          `json:"id"`
	Name                   string          `json:"name"`
	NumberOfVPCS           uint            `json:"number_of_vpcs"`
	NumberOfSubnetPerVpc   uint            `json:"number_of_subnet_per_vpc"`
	NumberOfSecurityGroups uint            `json:"number_of_security_groups"`
	SubnetCiders           []string        `json:"subnet_ciders"`
	Routers                []Router        `json:"routers"`
	Gateways               []Gateway       `json:"gateways"`
	SecurityGroups         []SecurityGroup `json:"security_groups"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

type Router struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	SubnetGateways []string `json:"subnet_gateways"`
}

type Gateway struct {
	Id   string   `json:"id"`
	Name string   `json:"name"`
	Ips  []string `json:"ips"`
}

type SecurityGroup struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Rules   []Rule   `json:"rules"`
	ApplyTo []string `json:"apply_to"`
}

type Rule struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	EtherType      string `json:"ethertype"`
	Direction      string `json:"direction"`
	Protocol       string `json:"protocol"`
	PortRange      string `json:"port_range"`
	RemoteGroupId  string `json:"remote_group_id"`
	RemoteIpPrefix string `json:"remote_ip_prefix"`
}

// Compute Configuration
type ComputeConfig struct {
	Id                       string           `json:"id"`
	Name                     string           `json:"name"`
	NumberOfComputeNodes     uint             `json:"number_of_compute_nodes"`
	NumberOfVmPerComputeNode uint             `json:"number_of_vm_per_compute_node"`
	NumberOfPortPerVm        uint             `json:"number_of_port_per_vm"`
	VmDeployType             string           `json:"vm_deploy_type"`
	Scheduler                string           `json:"scheduler"`
	NumberOfVpcs             uint             `json:"number_of_vpcs"`
	NumberOfVmPerVpc         []uint           `json:"number_of_vm_per_vpc"`
	NumberOfSubnetsPerVpc    []uint           `json:"number_of_subnets_per_vpc"`
	VPCInfo                  []ComputeVPCInfo `json:"vpc_info"`
	CreatedAt                time.Time        `json:"created_at"`
	UpdatedAt                time.Time        `json:"updated_at"`
}

type ComputeVPCInfo struct {
	VpcId           string              `json:"vpc_id"`
	NumberOfSubnets uint                `json:"number_of_subnets"`
	SubnetInfo      []ComputeSubnetInfo `json:"subnet_info"`
}

type ComputeSubnetInfo struct {
	SubnetId    string `json:"subnet_id"`
	NumberOfVMs uint   `json:"number_of_vms"`
}

// Test Configuration
type TestConfig struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Tests     []Test    `json:"tests"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Test struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	Script     string   `json:"script"`
	Cmd        string   `json:"cmd"`
	Parameters []string `json:"parameters"`
	WhenToRun  string   `json:"when_to_run"`
	WhereToRun string   `json:"where_to_run"`
}
