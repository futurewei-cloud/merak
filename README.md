# Merak

A Large-scale Cloud Emulator provides ability to
- emulate data center physical topologies and network devices including hosts, switches, and routers.
- emulate a large volume of compute nodes (more than 100K) with limited physical hardware resources.
- conduct the performance test for a target project's (e.g., Alcor) control plane with a large-size of VPC having more than 1M VMs.
- automatically create and conduct different performance test scenarios and collect results.

## Platforms

There are many different hardware resource management platform in the field, currently we choose two platforms to investigate and create our prototype:

- Kubernetes cluster with Meshnet CNI 
- Distrinet with LXD containers

## Architecture

The following diagram illustrate the high-level architecture of Merak on a kubernetes cluster using Meshnet CNI and the basic workflow to emulate Alcor's control plane for creating VMs in the emulated compute nodes.

![Merak Architecture](docs/images/merak-architecture.png)

### Components
- Scenario Manager: create the required topology and test scenarios.
- K8S-Topo: deploy pods with the given topology.
- Merak Network: create network infrastructure resources, e.g., vpcs, subnets, and security groups.
- Merak Compute: register compute nodes informantion, create VMs and collect test results from merak agents.
- Merak Agent: create virtual network devices (bridges, tap devices and veth pairs) and network namespace for VMs, collect test results and send the results back to merak compute.

## Scalability 
In order to provide more virtual and emulated resources with limited hardware resources, three possible solutions are investigated and developed in this project:
- Docker-in-Docker
- Kubernetes-in-Kubernetes (KinK)
- Kubernetes cluster in virtual machines

For ore detail design and information, please refer to the docs folder in this repository. 
