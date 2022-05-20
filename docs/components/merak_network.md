# Merak Network Design Document  

- [Merak Network Design Document](#merak-network-design-document)
  - [Overview](#overview)
  - [Service Requirements](#service-requirements)
  - [Design](#design)
  - [Data Schema](#data-schema)

## Overview  
___  
The Merak Network component is serving the purpose of creating modifying and deleting the Network configuration from the Scenario Manager. In our first version, it will support configuring network configurations with Alcor.  

## Service Requirements  
___  
1. Ability to create/modify/delete network configurations in Alcor  
2. Create Security Groupe  
3. Create/modify/delete Number of VPC 
4. Create/modify/delete Number of Subnet per VPC  
5. Create Router  
6. Connect Subnets to Router  
7. Store the above information in DB, and have the ability to delete them after each test run. Or have ability to modify our network configurations on the fly.  

## Design  
___  

- As for the first version of Merak Network component, it will mainly setup virtual network with [Alcor](https://github.com/futurewei-cloud/alcor).  
- To enable Merak Network have the ability to work with different SDN platform, such as [Alcor](https://github.com/futurewei-cloud/alcor). For now we will have different functions within the code to do different SDN network setup, later may switch to use different micro service do different SDN.  
- Once received request from the Merak Scenario Manager, Merak Network will start to do its job. Making VPC, Security Group, Router, Subnet; and attach them together.  
- After setup the virtual network, all information related to future use will be saved into DB. Currently planning on using the mongoDB, the schema will be show in the below section.  

![merak network diagram](../images/MerakNetworkFlow.drawio.svg)  

## Data Schema  
___  

For new we are planning on store data in Redis DB, below are the schema for each data point:  

- VPC:  
    ```  
    {
        "VPC": {
            "VPC_ID": "",
            "Name": "",
            "Project_id": "",
            "Tenant_id": "",
            "cider_size": "",
            "Tunnel_id": "",
            "security_group_id":"",
            "Gateways": ""
        }
    }
    ```  

- Security Group:  
    ```  
    {
        "security_group":{
            "security_group_id":"",
            "Project_id": "",
            "Tenant_id": ""
        }
    }
    ```  

- Router:  
    ```  
    {
        "Router": {
            "Router_ID": "",
            "Name": "",
            "Project_id": "",
            "Tenant_id": "",
            "security_group_id":"",
            "VPC_ID": ""
        }
    }
    ```  

- Subnet:  
    ```  
    {
        "Subnet": {
            "Subnet_ID": "",
            "Name": "",
            "Project_id": "",
            "Tenant_id": "",
            "cider_size": "",
            "VPC_ID": "",
            "security_group_id":"",
            "Router_ID": ""
        }
    }
    ```  

- SDN Path:  
  - This is an example for Alcor, will need further changes and more fields for all other SDN.  
    ```  
    {
        "SDN_path": {
            "Alcor": {
                "create_vpc_endpoint": "",
                "create_router_endpoint": "",
                "create_subnet_endpoint": "",
                "create_sg_endpoint": "",
                "attach_router_endpoint": ""
            }
        }
    }
    ```  
