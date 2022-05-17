# Merak Compute

Merak compute manages the creation and configuration of simulated virtual machines and endpoints.

![merak compute design diagram](images/merak_compute_design_diagram.png)

## Components

### Interface with Merak Scenario Manager

Merak compute watches API server for pods annotated by scenario manager.

Examples:
- meraksim.com/num_vms=5
- meraksim.com/endpoints_per_vm=6
- meraksim.com/test_icmp
- meraksim.com/test_tcp
- meraksim.com/test_udp

### Network Manager Plugin Interface

Need a runtime communication interface.
Possible Candidates:
gRPC
Kubernets CRD

###### gRPC client to plugin server.

Plugin server to be implemented by developers.
These plugins will be responsible for communicating with the network provider's control plane

#### Interface

Create VPC

- VNI
- ID
- IP
- Prefix
- Optional

Update VPC

- VNI
- ID
- IP
- Prefix
- Optional

Delete VPC

- VNI
- ID
- Optional

Create Subnet

- VNI
- ID
- VPC
- IP
- Prefix
- Optional

Update Subnet

- VNI
- ID
- VPC
- IP
- Prefix
- Optional

Delete Subnet

- VNI
- ID
- Optional

Create Node
- Name
- IP
- Mac
- Optional

Update Node
- Name
- IP
- Mac
- Optional

Delete Node
- Name
- IP
- Mac
- Optional


### Examples

#### Alcor

![merak network manager plugin alcor example diagram](/assets/images/merak_alcor_network_manager_plugin_example.png)

#### Mizar

![merak network manager plugin mizar example diagram](/assets/images/merak_mizar_network_manager_plugin_example.png)

# Merak Agent
Manages simulated VM creation and provides an interface to allocate an endpoint on a node for a specific network plugin.

Merak node images deployed as kubernetes pods.

#### Interface Between Merak Agent and Network Provider Plugin

Create Sim VM
- Name
- Optional

Update Sim VM
- Name
- Optional

Delete Sim VM
- Name
- Optional

Create Sim Endpoint
- ID
- IP
- MAC
- Subnet
- Interface Name
- Optional

Update Sim Endpoint
- ID
- IP
- MAC
- Subnet
- Interface Name
- Optional

Delete Sim Endpoint
- ID
- Optional

## Components

### Merak Node

![merak node design diagram](/assets/images/merak_node_design_diagram.png)

### Merak VM

![merak vm design diagram](/assets/images/merak_vm_design_diagram.png)

### Merak Endpoint

![merak endpoint design diagram](/assets/images/merak_endpoint_design_diagram.png)

### Examples

#### Alcor

Alcor Merak plugin will take the role of Openstack Neutron.

![merak alcor node example](/assets/images/merak_alcor_node_example.png)

#### Mizar

Mizar Merak plugin will take the role of local node pod operator.

![merak mizar node example](/assets/images/merak_mizar_node_example.png)

#### VM and Endpoint allocation plugin interface

Communicate to plugin via environment variables.
Network Provider dependent. Responsible for hooking up the dataplane devices and applications.

## End-to-end Workflow

### Alcor Example

![merak e2e alcor example](/assets/images/merak_e2e_alcor_example.png)

#### Monitoring/Data Collection

eBPF tools

#### Testing

iPerf inside network namespace