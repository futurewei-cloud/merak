#!/usr/bin/env python3
import subprocess
import requests
import json
import socket
import uuid
from argparse import ArgumentParser
from syslog import syslog
from time import sleep


def run_cmd(cmd):
    result = subprocess.Popen(
        cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    text = result.stdout.read().decode()
    returncode = result.returncode
    return (returncode, text)


def create_virtual_instance(namespace, ip, mac, prefix, outer_veth, inner_veth, bridge, tap, main_interface, subnet_ip):
    """
    Creates a veth pair.
    """

    script = (f''' bash -c '\
ip netns add {namespace} && \
ip link add {inner_veth} type veth peer name {outer_veth} && \
ip link set {inner_veth} netns {namespace} && \
ip netns exec {namespace} ip addr add {ip}/{prefix} dev {inner_veth} && \
ip netns exec {namespace} ip link set dev {inner_veth} up && \
ip netns exec {namespace} sysctl -w net.ipv4.tcp_mtu_probing=2 && \
ip link set dev {outer_veth} up && \
ip netns exec {namespace} ifconfig lo up &&  \
ip netns exec {namespace} ifconfig {inner_veth} hw ether {mac} && \
ip link add name {bridge} type bridge && \
ip link set {outer_veth} master {bridge} && \
ip link set {tap} master {bridge} && \
ip link set dev {bridge} up && \
ip link set dev {tap} up' ''')

    return_code, text = run_cmd(script)
    print(text)
    print(return_code)


def main():
    parser = ArgumentParser()
    description = "Parser for reading multiple subnets from command line arguments. Will use the first subnet if nothing is given"
    parser = ArgumentParser(description=description)
    cmd = "/root/alcor-control-agent/build/bin/AlcorControlAgent -d -a 10.213.43.112 -p 30014 > /dev/null 2>&1 &"
    text, returncode = run_cmd(cmd)
    print(text)
    print(returncode)
    syslog("ACA started")
    hostname = socket.gethostname()
    host_ip = socket.gethostbyname(hostname)
    syslog("Hostname is {} at IP {}".format(hostname, host_ip))
    address = "10.213.43.112"
    project_id = "123456789"
    port_name = "merak_port"
    inner_veth_name = "inner-" + uuid.uuid4().hex[-5:]
    outer_veth_name = "outer-" + uuid.uuid4().hex[-5:]
    netns = "ns-" + uuid.uuid4().hex[-5:]
    bridge_name = "br0-" + uuid.uuid4().hex[-5:]
    main_interface_name = "eth0"
    headers = {'Content-Type': 'application/json'}
    sm_port = "30002"
    pm_port = "30006"
    vpm_port = "30001"
    sgm_port = "30008"
    nmm_port = "30007"
    ncm_port = "30007"
    get_network_endpoint = "http://{}:{}/project/{}/subnets".format(
        address, sm_port, project_id)
    create_port_endpoint = "http://{}:{}/project/{}/ports".format(
        address, pm_port, project_id)
    get_sg_endpoint = "http://{}:{}/project/{}/security-groups".format(
        address, sgm_port, project_id)
    create_node_endpoint = "http://{}:{}/nodes".format(
        address, nmm_port)
    create_node_ncm_endpoint = "http://{}:{}/ncms".format(
        address, ncm_port)
    cmd = 'ip addr show ' + \
        "eth0" + \
        ' | grep "link/ether\\b" | awk \'{print $2}\' | cut -d/ -f1'
    r = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
    host_mac = r.stdout.read().decode().strip()
    # # Step 0:
    # # query for vpc a vpc
    #
    # get_vpc_endpoint = "http://{}:{}/project/{}/vpcs".format(
    #     address, vpm_port, project_id)
    # response = requests.get(get_vpc_endpoint)
    # while not response.ok:
    #     sleep(5)
    #     response = requests.get(get_vpc_endpoint)
    # vpc_id = response.json()["vpcs"][0]["id"]

    # Step 1:
    # Query Alcor about the network for the VM
    # Get tenant ID, and Network ID from response
    # Use the first subnet if no subnet is given

###############REGISTER NODE NCM###############
#     node_body = {
#         "ncm_info": {
#             "cap": 1,
#             "id": "ncm_id_1",
#             "uri": "ncm_uri_1"
#         }
#     }
#     syslog("###############REGISTER NCM NODE###############")
#     response = requests.post(create_node_ncm_endpoint,
#                              headers=headers,
#                              verify=False,
#                              data=json.dumps(node_body))
#     syslog("create_node_ncm response {}".format(response.text))
#     while not response.ok:
#         sleep(5)
#         syslog("create_node response {}".format(response.text))
#         response = requests.get(create_node_ncm_endpoint)
#     ncm_id = "ncm_id_1"
#     ncm_uri = "ncm_uri_1"
# ###############REGISTER NODE###############
#     node_body = {
#         "host_info": {
#             "host_dvr_mac": "string",
#             "local_ip": host_ip,
#             "mac_address": host_mac,
#             "ncm_id": ncm_id,
#             "ncm_uri": ncm_uri,
#             "node_id": hostname,
#             "node_name": hostname,
#             "server_port": 0,
#             "veth": "eth0"
#         }
#     }
#     syslog("###############REGISTER NODE###############")
#     response = requests.post(create_node_endpoint,
#                              headers=headers,
#                              verify=False,
#                              data=json.dumps(node_body))
#     syslog("create_node response {}".format(response.text))
#     while not response.ok:
#         sleep(5)
#         syslog("create_node response {}".format(response.text))
#         response = requests.get(create_node_endpoint)

###############GET SECURITY GROUP###############
    syslog("###############GET SECURITY GROUP###############")
    response = requests.get(get_sg_endpoint)
    syslog("get_sg response {}".format(response.text))
    while not response.ok:
        sleep(5)
        syslog("get_sg_endpoint response {}".format(response.text))
        response = requests.get(get_sg_endpoint)
    json_response = response.json()
    sg_id = json_response["security_groups"][0]["id"]
    tenant_id = json_response["security_groups"][0]["tenant_id"]

###############GET SUBNET###############

    syslog("###############GET SUBNET###############")
    syslog("get_subnet response {}".format(response.text))
    response = requests.get(get_network_endpoint)
    while not response.ok:
        sleep(5)
        syslog("get_subnet response {}".format(response.text))
        response = requests.get(get_network_endpoint)
    json_response = response.json()
    network_id = json_response["subnets"][0]["network_id"]
    # tenant_id = json_response["subnets"][0]["tenant_id"]
    prefix = json_response["subnets"][0]["cidr"].split("/")[1]
    subnet_id = json_response["subnets"][0]["id"]
    subnet_ip = json_response["subnets"][0]["cidr"].split("/")[0]

    parser.add_argument("-s", "--subnets", action="store", dest="subnets",
                        type=str, nargs="*", default=[subnet_id],
                        help="Examples: -i subnet1, subnet2, subnet3")
    opts = parser.parse_args()

    for subnet in opts.subnets:
        print("Creating VM in subnet: {}".format(subnet))

    ###############CREATE MINIMAL PORT###############
        create_minimal_port_body = {
            "port": {
                "admin_state_up": True,
                "device_id": netns,
                "network_id": network_id,
                "security_groups": [
                    sg_id
                ],
                "fixed_ips": [
                    {
                        "subnet_id": subnet
                    }
                ],
                "tenant_id": tenant_id
            }
        }
        syslog("###############CREATE MINIMAL PORT###############")
        response = requests.post(create_port_endpoint,
                                 headers=headers,
                                 verify=False,
                                 data=json.dumps(create_minimal_port_body))
        syslog("create_port response {}".format(response.text))
        while not response.ok:
            sleep(5)
            syslog("create_port response {}".format(response.text))
            response = requests.post(create_port_endpoint,
                                     headers=headers,
                                     verify=False,
                                     data=json.dumps(create_minimal_port_body))
        json_response = response.json()
        ip = json_response["port"]["fixed_ips"][0]["ip_address"]
        mac = json_response["port"]["mac_address"]
        tap_name = "tap" + json_response["port"]["id"][:11]
        port_id = json_response["port"]["id"]

    ###############CREATE VM###############
        syslog("###############CREATE VM###############")
        create_virtual_instance(
            netns, ip, mac, prefix, outer_veth_name, inner_veth_name, bridge_name, tap_name, main_interface_name, subnet_ip)

    ###############UPDATE PORT###############
        update_port_body = {
            "port": {
                "project_id": project_id,
                "id": port_id,
                "name": port_name,
                "description": "",
                "network_id": network_id,
                "tenant_id": tenant_id,
                "admin_state_up": True,
                "veth_name": inner_veth_name,
                "device_id": hostname,
                "device_owner": "compute:nova",
                "fast_path": True,
                "binding:host_id": hostname
            }
        }
        syslog("###############UPDATE PORT###############")
        update_port_endpoint = "http://{}:{}/project/{}/ports/{}".format(
            address, pm_port, project_id, port_id)
        response = requests.put(update_port_endpoint, headers=headers, verify=False,
                                data=json.dumps(update_port_body))
        syslog("update_port response {}".format(response.text))
        while not response.ok:
            sleep(5)
            syslog("update_port response {}".format(response.text))
            response = requests.put(update_port_endpoint, headers=headers, verify=False,
                                    data=json.dumps(update_port_body))


main()
