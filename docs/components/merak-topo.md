# Overview
This is a page for merak-topo design document.
Merak-topo is a Merak function which create/update/delete network topology in the Kubernete cluster based on the request from the Merak Scenario Manager.

# Service Requriements
<ol>
1. Pull the test target image (ACA+OVS, Mizar)
<ol>
    1.1. Save the targe image <br>
    1.2: Rebuild image by adding golang environment and install Merak Agent <br>
</ol>
2. Run Merak Agent during the pod deploying <br>
3. Create gRPC channel with Scenario Manager <br>
4. Parse the protobuf message from Scenario Manger <br>
5. Create/delete the topology <br>
6. Update the topology <br>
</ol>

# Design
In order to communicate with the Scenario Manager and operate on topology in the Kubernete cluster, the main resource for the Merak-topo is designed as the following workflow, data schema, and design diagram.

## Workflow
This is the main workflows of Merak-topo based on the received message from the Scenario Manager, including the operation for creating, deleting, and updating a topology.
### Create 
![merak-topo create topology workflow](../images/merak-topo_create_topology_workflow.png)


### Delete 
![merak-topo delete topology workflow](../images/merak-topo_delete_topology_workflow.png)


### Update 
![merak-topo update topology workflow](../images/merak-topo_create_topology_workflow.png)


## Data Schema

