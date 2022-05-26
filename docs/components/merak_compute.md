# Merak Compute

Merak compute manages the allocation of virtual machines and ports.

![merak compute design diagram](../images/merak_compute_design_diagram.png)

## Services

The following services are provided over gRPC

Compute Scenarios:
- INFO
  - Returns information about the current status of the scheduled VMs and Ports.
- CREATE
  - Creates a new set of VMs and Ports.
- UPDATE
  - Update an existing set of VMs and Ports.
- DELETE
  - Delete an existing set of VMs and Ports.

Test Scenarios:
- INFO
  - Returns information about the status of an existing test scenario
- CREATE
  - Creates a new test scenario
- UPDATE
  - Update an existing test scenario
- DELETE
  - Delete an existing test scenario

## Components

### Merak Compute Controller
The Merak Compute Controller will be responsible for receiving, parsing, and acting on requests sent  from the scenario manager.
The compute manager will also update the database with all schedulable pods.

### VM Controller
The VM Controller will be responsible for making calls to the Merak Agent to Create/Update/Delete VMs.
#### Interface with Merak Agent

get_vm(hostname, vm):
- Returns info about the VM on a node.

get_vm_node(hostname):
- Returns info about all VMs on the node.

create_n_vms_on_host(hostname, n)
- creates n VMs at hostname
- returns a list of names of VMs created

delete_n_vms(hostname)
- deletes n VMs at hostname
- returns a list of names of VMs deleted

### Port Controller
The Port Controller will be responsible for making calls to the Merak Agent to Create/Update/Delete Ports.

#### Interface with Merak Agent

get_ports_vm(hostname, vm):
- Returns info on all ports in the VM on the node.

get_ports_node(hostname):
- Returns info on all ports on the node.

create_n_ports(hostname, vm, tenant, vpc, subnet, security_group)
- creates n ports in vm at hostname in the described VPC and subnet
- returns a list of names and IP of the ports created

delete_n_ports(hostname, vm)
- deletes n ports at hostname in vm
- returns a list of names of ports deleted

### Test Controller

The Test controller will be responsible for coordinating tests across the available vms.

#### Interface with Merak Agent

get_test(vm, src):
- Returns the status of a running test on the VM origination from the source

run_test(vm, src, target, test-type, opt):
- Runs a network test from inside the VM to the target
- Returns the result of the ping test

stop_test(vm, src):
- Stops any running test in the VM originating from the source

## Scheduling

Merak Compute will assume that the Kubernetes scheduler has uniformally distributed its pods across all nodes in the cluster.

#### VM/Port Distribution

The following are the four VM and port distribution settings.

**Manual**: VMs/Ports are manually assigned to an existing port.

**Random**: Schedules VMs/Pods randomly

**Skew**: Schedule majority of VM/Pods on a small group of hosts.

**Uniform**: Schedule all VM/Pods evenly.

#### VM/Port Schedule Rate

The following are the three VM and Port scheduling settings.

**Sequential**: Each VM/Port will be created one-by-one.

**RPS**: VMs/Ports will be created at a given rate given by the Scenario Manager.

**Random**: VM/Port will be created at a random rate.

## Data Model



#### Compute Datamodel

- Pod
  - ID
  - IP
    - VMs

- VM
  - ID
    - Ports

- Ports
  - ID
  - VM
  - Tenant
  - VPC
  - Subnet
  - Security Group
  - IP


Example:
```
{
    "pod":
    {
        "id": "pod1",
        "ip": "10.0.0.2",
        "vms": ["vm1","vm2"]
    }
}
```

```
{
  "vm":
  {
      "id": "vm1"
      "ports": ["port1", "port2"]
  }
}
```

```
{
  "port":
  {
      "id": "port1",
      "vm": "vm1"
      "tenant": "tenant1",
      "vpc": "vpc1",
      "subnet": "subnet1",
      "security_group": "sg1",
      "ip": "20.0.0.2",
  }
}

```

#### Test Datamodel

- Test
  - Name
  - source
    - host
      - vm
        - port
  - target
    - host
      - vm
        - port
  - test-type
  - results

Example:
```
{
    test:
    {
        "name": "test-ping",
        "source":
        [
            {
                "host": "pod1",
                "vm":   "vm1",
                "port": "port1"
            },
            {
                "host": "pod1",
                "vm":   "vm1",
                "port": "port2"
            },

        ],
        "target":
        [
            {
                "host": "pod2",
                "vm":   "vm1",
                "port": "port1"
            },
            {
                "host": "pod2",
                "vm":   "vm2",
                "port": "port1"
            },

        ],
        "test-type": "ping",
        "results":
        [
            "source_1->target_1": "pass",
            "source1->target_2": "failed",
            "source2->target_1": "pass",
            "source2->target_2": "pending"
        ]

    }
}
```
### Data Storage
Merak Compute will use a distributed KV Datastore behind a Kubernetes ClusterIP service.
