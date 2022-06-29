package entities

import (
	"time"
)

type AppConfig struct {
	DBHost    string `yaml:"db_host"`
	DBPort    string `yaml:"db_port"`
	DBUser    string `yaml:"db_user"`
	DBPass    string `yaml:"db_pass"`
	LogLevel  string `yaml:"log_level"`
	UseSyslog bool   `yaml:"use_syslog"`
}

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

type EventName string

const (
	EVENT_DEPLOY EventName = "DEPLOY"
	EVENT_DELETE EventName = "DELETE"
	EVENT_UPDATE EventName = "UPDATE"
	EVENT_CHECK  EventName = "CHECK"
)

// Action
type ScenarioAction struct {
	ScenarioId string          `json:"scenario_id"`
	Services   []ServiceAction `json:"services"`
}

type ServiceAction struct {
	ServiceName string        `json:"service_name"`
	Action      EventName     `json:"action"`
	Status      ServiceStatus `json:"status"`
}

// Scenario
type Scenario struct {
	Id            string        `json:"id" swaggerignore:"true"`
	Name          string        `json:"name"`
	ProjectId     string        `json:"project_id"`
	TopologyId    string        `json:"topology_id"`
	ServiceConfId string        `json:"service_config_id"`
	NetworkConfId string        `json:"network_config_id"`
	ComputeConfId string        `json:"compute_config_id"`
	TestConfId    string        `json:"test_config_id"`
	Status        ServiceStatus `json:"status" swaggerignore:"true"`
	CreatedAt     time.Time     `json:"created_at" swaggerignore:"true"`
	UpdatedAt     time.Time     `json:"updated_at" swaggerignore:"true"`
}

// Service Configuration
type ServiceConfig struct {
	Id        string    `json:"id" swaggerignore:"true"`
	Name      string    `json:"name"`
	Services  []Service `json:"services"`
	CreatedAt time.Time `json:"created_at" swaggerignore:"true"`
	UpdatedAt time.Time `json:"updated_at" swaggerignore:"true"`
}

type Service struct {
	Id           string   `json:"id" swaggerignore:"true"`
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
	Id               string        `json:"id" swaggerignore:"true"`
	Name             string        `json:"name"`
	TopoType         string        `json:"type"`
	NumberOfVhosts   uint          `json:"number_of_vhosts"`
	NumberOfRacks    uint          `json:"number_of_racks"`
	VhostsPerRack    uint          `json:"vhosts_per_rack"`
	DataPlaneCidr    string        `json:"data_plane_cidr"`
	NumberOfGateways uint          `json:"number_of_control_plane_gateways"`
	GatewayIPs       []string      `json:"control_plane_gateway_ips"`
	Images           []Image       `json:"images"`
	VNodes           []VNode       `json:"vnodes"`
	VLinks           []VLink       `json:"vlinks"`
	Status           ServiceStatus `json:"status" swaggerignore:"true"`
	CreatedAt        time.Time     `json:"created_at" swaggerignore:"true"`
	UpdatedAt        time.Time     `json:"updated_at" swaggerignore:"true"`
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
	Id                     string          `json:"id" swaggerignore:"true"`
	Name                   string          `json:"name"`
	NumberOfVPCS           uint            `json:"number_of_vpcs"`
	NumberOfSubnetPerVpc   uint            `json:"number_of_subnet_per_vpc"`
	NumberOfSecurityGroups uint            `json:"number_of_security_groups"`
	Vpcs                   []VPCInfo       `json:"vpcs"`
	Routers                []Router        `json:"routers"`
	Gateways               []Gateway       `json:"gateways"`
	SecurityGroups         []SecurityGroup `json:"security_groups"`
	Status                 ServiceStatus   `json:"status" swaggerignore:"true"`
	CreatedAt              time.Time       `json:"created_at" swaggerignore:"true"`
	UpdatedAt              time.Time       `json:"updated_at" swaggerignore:"true"`
}

type Router struct {
	Id             string   `json:"id" swaggerignore:"true"`
	Name           string   `json:"name"`
	SubnetGateways []string `json:"subnet_gateways"`
}

type Gateway struct {
	Id   string   `json:"id" swaggerignore:"true"`
	Name string   `json:"name"`
	Ips  []string `json:"ips"`
}

type SecurityGroup struct {
	Id        string   `json:"id" swaggerignore:"true"`
	Name      string   `json:"name"`
	TenantId  string   `json:"tenant_id"`
	ProjectId string   `json:"project_id"`
	Rules     []Rule   `json:"rules"`
	ApplyTo   []string `json:"apply_to"`
}

type Rule struct {
	Id             string `json:"id" swaggerignore:"true"`
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
	Id                   string        `json:"id" swaggerignore:"true"`
	Name                 string        `json:"name"`
	NumberOfComputeNodes uint          `json:"number_of_compute_nodes"`
	NumberOfPortPerVm    uint          `json:"number_of_port_per_vm"`
	VmDeployType         string        `json:"vm_deploy_type"`
	Scheduler            string        `json:"scheduler"`
	NumberOfVmPerVpc     uint          `json:"number_of_vm_per_vpc"`
	VPCInfo              []VPCInfo     `json:"vpc_info"`
	Status               ServiceStatus `json:"status" swaggerignore:"true"`
	CreatedAt            time.Time     `json:"created_at" swaggerignore:"true"`
	UpdatedAt            time.Time     `json:"updated_at" swaggerignore:"true"`
}

type VPCInfo struct {
	VpcId           string       `json:"vpc_id" swaggerignore:"true"`
	TenantId        string       `json:"tenant_id"`
	ProjectId       string       `json:"project_id"`
	VpcCidr         string       `json:"vpc_cidr"`
	NumberOfSubnets uint         `json:"number_of_subnets"`
	SubnetInfo      []SubnetInfo `json:"subnet_info"`
}

type SubnetInfo struct {
	SubnetId      string `json:"subnet_id" swaggerignore:"true"`
	SubnetCidr    string `json:"subnet_cidr"`
	SubnetGateway string `json:"subnet_gateway"`
	NumberOfVMs   uint   `json:"number_of_vms"`
}

// Test Configuration
type TestConfig struct {
	Id        string        `json:"id" swaggerignore:"true"`
	Name      string        `json:"name"`
	Tests     []Test        `json:"tests"`
	Status    ServiceStatus `json:"status" swaggerignore:"true"`
	CreatedAt time.Time     `json:"created_at" swaggerignore:"true"`
	UpdatedAt time.Time     `json:"updated_at" swaggerignore:"true"`
}

type Test struct {
	Id         string   `json:"id" swaggerignore:"true"`
	Name       string   `json:"name"`
	Script     string   `json:"script"`
	Cmd        string   `json:"cmd"`
	Parameters []string `json:"parameters"`
	WhenToRun  string   `json:"when_to_run"`
	WhereToRun string   `json:"where_to_run"`
}
