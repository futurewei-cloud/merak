#!/usr/bin/env python3
import requests
import json
from time import sleep

def main():
    address = "10.213.43.224"
    project_id = "123456789"
    router_id = "11112801-d675-4688-a63f-dcda8d327f50"
    ncm_port = "30007"
    sgm_port = "30008"
    vpm_port = "30001"
    sm_port = "30002"
    router_port = "30003"
    headers = {'Content-Type': 'application/json'}
    create_node_ncm_endpoint = "http://{}:{}/ncms".format(
        address, ncm_port)
    create_sg_endpoint = "http://{}:{}/project/{}/security-groups".format(
        address, sgm_port, project_id)
    create_subnet_endpoint = "http://{}:{}/project/{}/subnets".format(
        address, sm_port, project_id)
    create_vpc_endpoint = "http://{}:{}/project/{}/vpcs".format(
         address, vpm_port, project_id)
    create_router_endpoint = "http://{}:{}/project/{}/routers".format(
         address, router_port, project_id)
    attach_router_endpoint = "http://{}:{}/project/{}/routers/{}/add_router_interface".format(
         address, router_port, project_id, router_id)

    # print("###############REGISTER NODE NCM###############")
    # ncm_body = {
    #     "ncm_info": {
    #         "cap": 1,
    #         "id": "ncm_id_1",
    #         "uri": "ncm_uri_1"
    #     }
    # }
    # response = requests.post(create_node_ncm_endpoint,
    #                          headers=headers,
    #                          verify=False,
    #                          data=json.dumps(ncm_body))
    # print("create_ncm response {}".format(response.text))
    # while not response.ok:
    #     sleep(5)
    #     response = requests.get(create_node_ncm_endpoint)

    print("###############CREATE VPC###############")
    network_body = {
        "network": {
            "admin_state_up": True,
            "revision_number": 0,
            "cidr": "10.0.0.0/16",
            "default": True,
            "description": "vpc",
            "dns_domain": "domain",
            "id": "9192a4d4-ffff-4ece-b3f0-8d36e3d88001",
            "is_default": True,
            "mtu": 1400,
            "name": "sample_vpc",
            "port_security_enabled": True,
            "project_id": "123456789"
        }
    }

    response = requests.post(create_vpc_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(network_body))
    print("create_vpc response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("create_vpc response {}".format(response.text))
        response = requests.get(create_vpc_endpoint)

    print("###############CREATE SG###############")

    sg_body = {
        "security_group": {
            "create_at": "string",
            "description": "string",
            "id": "3dda2801-d675-4688-a63f-dcda8d111111",
            "name": "sg1",
            "project_id": "123456789",
            "security_group_rules": [
            ],
            "tenant_id": "123456789",
            "update_at": "string"
        }
    }

    response = requests.post(create_sg_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(sg_body))
    print("create_sg response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("create_sg response {}".format(response.text))
        response = requests.get(create_sg_endpoint)

    print("###############CREATE SUBNET1###############")

    subnet_body = {
        "subnet":
        {
            "cidr": "10.0.1.0/24",
            "id": "8182a4d4-ffff-4ece-b3f0-8d36e3d88001",
            "ip_version": 4,
            "network_id": "9192a4d4-ffff-4ece-b3f0-8d36e3d88001",
            "name": "subnet1"
        }
    }
    response = requests.post(create_subnet_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(subnet_body))
    print("create_subnet response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("create_subnet response {}".format(response.text))
        response = requests.get(create_subnet_endpoint)

    print("###############CREATE SUBNET2###############")

    subnet_body = {
        "subnet":
        {
            "cidr": "10.0.2.0/24",
            "id": "8182a4d4-ffff-4ece-b3f0-8d36e3d88002",
            "ip_version": 4,
            "network_id": "9192a4d4-ffff-4ece-b3f0-8d36e3d88001",
            "name": "subnet2"
        }
    }
    response = requests.post(create_subnet_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(subnet_body))
    print("create_subnet response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("create_subnet response {}".format(response.text))
        response = requests.get(create_subnet_endpoint)

    print("###############CREATE SUBNET3###############")

    subnet_body = {
        "subnet":
        {
            "cidr": "10.0.3.0/24",
            "id": "8182a4d4-ffff-4ece-b3f0-8d36e3d88003",
            "ip_version": 4,
            "network_id": "9192a4d4-ffff-4ece-b3f0-8d36e3d88001",
            "name": "subnet2"
        }
    }
    response = requests.post(create_subnet_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(subnet_body))
    print("create_subnet response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("create_subnet response {}".format(response.text))
        response = requests.get(create_subnet_endpoint)

    print("###############CREATE SUBNET4###############")

    subnet_body = {
        "subnet":
        {
            "cidr": "10.0.4.0/24",
            "id": "8182a4d4-ffff-4ece-b3f0-8d36e3d88004",
            "ip_version": 4,
            "network_id": "9192a4d4-ffff-4ece-b3f0-8d36e3d88001",
            "name": "subnet2"
        }
    }
    response = requests.post(create_subnet_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(subnet_body))
    print("create_subnet response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("create_subnet response {}".format(response.text))
        response = requests.get(create_subnet_endpoint)

    print("###############CREATE SUBNET5###############")

    subnet_body = {
        "subnet":
        {
            "cidr": "10.0.5.0/24",
            "id": "8182a4d4-ffff-4ece-b3f0-8d36e3d88005",
            "ip_version": 4,
            "network_id": "9192a4d4-ffff-4ece-b3f0-8d36e3d88001",
            "name": "subnet2"
        }
    }
    response = requests.post(create_subnet_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(subnet_body))
    print("create_subnet response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("create_subnet response {}".format(response.text))
        response = requests.get(create_subnet_endpoint)

    print("###############CREATE ROUTER###############")

    router_body = {
        "router": {
            "admin_state_up": True,
            "availability_zone_hints": [
            "string"
            ],
            "availability_zones": [
            "string"
            ],
            "conntrack_helpers": [
            "string"
            ],
            "description": "string",
            "distributed": True,
            "external_gateway_info": {
            "enable_snat": True,
            "external_fixed_ips": [],
            "network_id": "9192a4d4-ffff-4ece-b3f0-8d36e3d88001"
            },
            "flavor_id": "string",
            "gateway_ports": [
            ],
            "ha": True,
            "id": "11112801-d675-4688-a63f-dcda8d327f50",
            "name": "router1",
            "owner": "string",
            "project_id": "123456789",
            "revision_number": 0,
            "routetable": {},
            "service_type_id": "string",
            "status": "BUILD",
            "tags": [
            "string"
            ],
            "tenant_id": "123456789"
        }
    }
    response = requests.post(create_router_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(router_body))
    print("create_router response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("create_router response {}".format(response.text))
        response = requests.get(create_router_endpoint)


    print("###############ATTACH SUBNET1 TO ROUTER###############")
    attachRouter_body = {
        "subnet_id": "8182a4d4-ffff-4ece-b3f0-8d36e3d88001"
    }
    response = requests.put(attach_router_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(attachRouter_body))
    print("attachRouter response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("attachRouter response {}".format(response.text))
        response = requests.get(attach_router_endpoint)

    print("###############ATTACH SUBNET2 TO ROUTER###############")
    attachRouter_body = {
        "subnet_id": "8182a4d4-ffff-4ece-b3f0-8d36e3d88002"
    }
    response = requests.put(attach_router_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(attachRouter_body))
    print("attachRouter response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("attachRouter response {}".format(response.text))
        response = requests.get(attach_router_endpoint)

    print("###############ATTACH SUBNET3 TO ROUTER###############")
    attachRouter_body = {
        "subnet_id": "8182a4d4-ffff-4ece-b3f0-8d36e3d88003"
    }
    response = requests.put(attach_router_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(attachRouter_body))
    print("attachRouter response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("attachRouter response {}".format(response.text))
        response = requests.get(attach_router_endpoint)

    print("###############ATTACH SUBNET4 TO ROUTER###############")
    attachRouter_body = {
        "subnet_id": "8182a4d4-ffff-4ece-b3f0-8d36e3d88004"
    }
    response = requests.put(attach_router_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(attachRouter_body))
    print("attachRouter response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("attachRouter response {}".format(response.text))
        response = requests.get(attach_router_endpoint)

    print("###############ATTACH SUBNET5 TO ROUTER###############")
    attachRouter_body = {
        "subnet_id": "8182a4d4-ffff-4ece-b3f0-8d36e3d88005"
    }
    response = requests.put(attach_router_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(attachRouter_body))
    print("attachRouter response {}".format(response.text))
    while not response.ok:
        sleep(5)
        print("attachRouter response {}".format(response.text))
        response = requests.get(attach_router_endpoint)

main()