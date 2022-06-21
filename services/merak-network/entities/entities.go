package entities

type VpcBody struct {
	AdminStateUp        bool   `json:"admin_state_up"`
	RevisionNumber      int    `json:"revision_number"`
	Cidr                string `json:"cidr"`
	ByDefault           bool   `json:"default"`
	Description         string `json:"description"`
	DnsDomain           string `json:"dns_domain"`
	Id                  string `json:"id"`
	IsDefault           bool   `json:"is_default"`
	Mtu                 int    `json:"mtu"`
	Name                string `json:"name"`
	PortSecurityEnabled bool   `json:"port_security_enabled"`
	ProjectId           string `json:"project_id"`
}
type VpcStruct struct {
	Network VpcBody `json:"network"`
}
type VpcReturn struct {
	Network struct {
		Default     bool   `json:"default"`
		ID          string `json:"id"`
		ProjectID   string `json:"project_id"`
		TenantID    string `json:"tenant_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Cidr        string `json:"cidr"`
		Routes      []struct {
			Destination       string `json:"destination"`
			Target            string `json:"target"`
			Priority          int    `json:"priority"`
			AssociatedType    string `json:"associatedType"`
			AssociatedTableID string `json:"associatedTableId"`
			ID                string `json:"id"`
			ProjectID         string `json:"project_id"`
			TenantID          string `json:"tenant_id"`
			Name              string `json:"name"`
			Description       string `json:"description"`
		} `json:"routes"`
		Router                  interface{}   `json:"router"`
		AdminStateUp            bool          `json:"admin_state_up"`
		DNSDomain               string        `json:"dns_domain"`
		Mtu                     int           `json:"mtu"`
		PortSecurityEnabled     bool          `json:"port_security_enabled"`
		ProviderNetworkType     string        `json:"provider:network_type"`
		ProviderPhysicalNetwork interface{}   `json:"provider:physical_network"`
		ProviderSegmentationID  int           `json:"provider:segmentation_id"`
		RouterExternal          bool          `json:"router:external"`
		Segments                interface{}   `json:"segments"`
		Shared                  bool          `json:"shared"`
		VlanTransparent         bool          `json:"vlan_transparent"`
		IsDefault               bool          `json:"is_default"`
		AvailabilityZoneHints   interface{}   `json:"availability_zone_hints"`
		AvailabilityZones       []string      `json:"availability_zones"`
		QosPolicyID             interface{}   `json:"qos_policy_id"`
		RevisionNumber          int           `json:"revision_number"`
		Status                  string        `json:"status"`
		Tags                    []interface{} `json:"tags"`
		CreatedAt               string        `json:"created_at"`
		UpdatedAt               string        `json:"updated_at"`
		Ipv4AddressScope        interface{}   `json:"ipv4_address_scope"`
		Ipv6AddressScope        interface{}   `json:"ipv6_address_scope"`
		L2Adjacency             interface{}   `json:"l2_adjacency"`
		Subnets                 interface{}   `json:"subnets"`
	} `json:"network"`
}

type SgBody struct {
	CreateAt           string   `json:"create_at"`
	Description        string   `json:"description"`
	Id                 string   `json:"id"`
	Name               string   `json:"name"`
	ProjectId          string   `json:"project_id"`
	SecurityGroupRules []string `json:"security_group_rules"`
	TenantId           string   `json:"tenant_id"`
	UpdateAt           string   `json:"update_at"`
}
type SgStruct struct {
	Sg SgBody `json:"security_group"`
}

type SubnetBody struct {
	Cider     string `json:"cidr"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	IpVersion int    `json:"ip_version"`
	NetworkId string `json:"network_id"`
}
type SubnetStruct struct {
	Subnet SubnetBody `json:"subnet"`
}
type SubnetReturn struct {
	Subnet struct {
		ID                string      `json:"id"`
		ProjectID         string      `json:"project_id"`
		TenantID          string      `json:"tenant_id"`
		Name              string      `json:"name"`
		Description       interface{} `json:"description"`
		NetworkID         string      `json:"network_id"`
		Cidr              string      `json:"cidr"`
		AvailabilityZone  interface{} `json:"availability_zone"`
		GatewayIP         string      `json:"gateway_ip"`
		GatewayPortID     string      `json:"gatewayPortId"`
		GatewayPortDetail struct {
			GatewayMacAddress string `json:"gateway_macAddress"`
			GatewayPortID     string `json:"gateway_port_id"`
		} `json:"gateway_port_detail"`
		AttachedRouterID  interface{}   `json:"attached_router_id"`
		PortDetail        interface{}   `json:"port_detail"`
		EnableDhcp        bool          `json:"enable_dhcp"`
		PrimaryDNS        interface{}   `json:"primary_dns"`
		SecondaryDNS      interface{}   `json:"secondary_dns"`
		DNSList           interface{}   `json:"dns_list"`
		IPVersion         int           `json:"ip_version"`
		IPV4RangeID       string        `json:"ipV4_rangeId"`
		IPV6RangeID       interface{}   `json:"ipV6_rangeId"`
		Ipv6AddressMode   interface{}   `json:"ipv6_address_mode"`
		Ipv6RaMode        interface{}   `json:"ipv6_ra_mode"`
		RevisionNumber    int           `json:"revision_number"`
		SegmentID         interface{}   `json:"segment_id"`
		Shared            interface{}   `json:"shared"`
		SortDir           interface{}   `json:"sort_dir"`
		SortKey           interface{}   `json:"sort_key"`
		SubnetpoolID      interface{}   `json:"subnetpool_id"`
		DNSPublishFixedIP bool          `json:"dns_publish_fixed_ip"`
		Tags              []interface{} `json:"tags"`
		TagsAny           interface{}   `json:"tags-any"`
		NotTags           interface{}   `json:"not-tags"`
		NotTagsAny        interface{}   `json:"not-tags-any"`
		Fields            interface{}   `json:"fields"`
		DNSNameservers    []interface{} `json:"dns_nameservers"`
		AllocationPools   []struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"allocation_pools"`
		HostRoutes           []interface{} `json:"host_routes"`
		Prefixlen            interface{}   `json:"prefixlen"`
		UseDefaultSubnetPool bool          `json:"use_default_subnet_pool"`
		ServiceTypes         []interface{} `json:"service_types"`
		CreatedAt            string        `json:"created_at"`
		UpdatedAt            string        `json:"updated_at"`
	} `json:"subnet"`
}

type RouterExternalGatewayInfo struct {
	EnableSnat       bool     `json:"enable_snat"`
	ExternalFixedIps []string `json:"external_fixed_ips"`
	NetworkId        string   `json:"network_id"`
}
type RouterRouterTable struct {
}
type RouterBody struct {
	AdminStateUp          bool                      `json:"admin_state_up"`
	AvailabilityZoneHints []string                  `json:"availability_zone_hints"`
	AvailiabilityZones    []string                  `json:"availability_zones"`
	ContrackHelpers       []string                  `json:"conntrack_helpers"`
	Description           string                    `json:"description"`
	Distributed           bool                      `json:"distributed"`
	ExternalGatewayInfo   RouterExternalGatewayInfo `json:"external_gateway_info"`
	FlavorId              string                    `json:"flavor_id"`
	GatewayPorts          []string                  `json:"gateway_ports"`
	Ha                    bool                      `json:"ha"`
	Id                    string                    `json:"id"`
	Name                  string                    `json:"name"`
	Owner                 string                    `json:"owner"`
	ProjectId             string                    `json:"project_id"`
	RevisionNumber        int                       `json:"revision_number"`
	RouterTable           RouterRouterTable         `json:"routetable"`
	ServiceTypeId         string                    `json:"service_type_id"`
	Status                string                    `json:"status"`
	Tags                  []string                  `json:"tags"`
	TenantId              string                    `json:"tenant_id":`
}
type RouterStruct struct {
	Router RouterBody `json:"router"`
}
type RouterReturn struct {
	Router struct {
		ID          string `json:"id"`
		ProjectID   string `json:"project_id"`
		TenantID    string `json:"tenant_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Routetable  struct {
			ID             string        `json:"id"`
			ProjectID      string        `json:"project_id"`
			TenantID       string        `json:"tenant_id"`
			Name           string        `json:"name"`
			Description    string        `json:"description"`
			Routes         []interface{} `json:"routes"`
			RouteTableType string        `json:"route_table_type"`
			Owner          string        `json:"owner"`
		} `json:"routetable"`
		Owner               string      `json:"owner"`
		GatewayPorts        interface{} `json:"gateway_ports"`
		AdminStateUp        bool        `json:"admin_state_up"`
		Status              string      `json:"status"`
		ExternalGatewayInfo struct {
			NetworkID        string      `json:"network_id"`
			EnableSnat       bool        `json:"enable_snat"`
			ExternalFixedIps interface{} `json:"external_fixed_ips"`
		} `json:"external_gateway_info"`
		RevisionNumber        int           `json:"revision_number"`
		Distributed           bool          `json:"distributed"`
		Ha                    bool          `json:"ha"`
		AvailabilityZoneHints interface{}   `json:"availability_zone_hints"`
		AvailabilityZones     []string      `json:"availability_zones"`
		ServiceTypeID         string        `json:"service_type_id"`
		FlavorID              string        `json:"flavor_id"`
		Tags                  []interface{} `json:"tags"`
		ConntrackHelpers      []interface{} `json:"conntrack_helpers"`
	} `json:"router"`
}

type AttachRouterStruct struct {
	SubnetId string `json:"subnet_id"`
}
type AttachRouterReturn struct {
	ID        string        `json:"id"`
	NetworkID string        `json:"network_id"`
	PortID    string        `json:"port_id"`
	SubnetID  string        `json:"subnet_id"`
	SubnetIds []string      `json:"subnet_ids"`
	ProjectID string        `json:"project_id"`
	TenantID  string        `json:"tenant_id"`
	Tags      []interface{} `json:"tags"`
}
