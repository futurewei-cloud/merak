#!/usr/bin/env python3
from platform import node
import subprocess
import requests
import json
import socket
import uuid
from syslog import syslog
from time import sleep


def run_cmd(cmd):
    result = subprocess.Popen(
        cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    text = result.stdout.read().decode()
    returncode = result.returncode
    return (returncode, text)

def main():

    # list all aca
    print("list all aca")
    cmd = "kubectl get pods"
    returncode, text = run_cmd(cmd)
    # print(returncode)
    print("text:", text)

    podNames = []
    lines = text.splitlines()
    for line in lines:
        if "aca-" in line:
            # print(line)
            podName = line.split()[0]
            podNames.append(podName)
    print(podNames)

    nodes_body = { "host_infos": []}

    for podName in podNames:

        cmd = "kubectl exec -it " + podName + " -- /bin/bash -c ifconfig"
        returncode, text = run_cmd(cmd)
        # print(returncode)
        # print(text)

        lines = text.splitlines()

        # for eth0
        n = 0
        foundText = False

        podEth0Ip = ""
        podEth0Mac = ""

        for line in lines:
            # print(line)
            if foundText == True:
                if n == 0:
                    podEth0Ip = line.split()[1]
                    print(podEth0Ip)
                if n == 1:
                    podEth0Mac = line.split()[1]
                    print(podEth0Mac)
                n+=1
            if "eth0" in line:
                foundText = True
            if n > 1:
                break

        # for eth1
        n = 0
        foundText = False

        podEth1Ip = ""
        podEth1Mac = ""

        for line in lines:
            # print(line)
            if foundText == True:
                if n == 0:
                    podEth1Ip = line.split()[1]
                    print(podEth1Ip)
                if n == 1:
                    podEth1Mac = line.split()[1]
                    print(podEth1Mac)
                n+=1
            if "eth1" in line:
                foundText = True
            if n > 1:
                break

        node_body = {
            # "host_dvr_mac": podEth1Mac,
            "local_ip": podEth1Ip,
            "mac_address": podEth1Mac,
            # "ncm_id": "ncm_id_1",
            # "ncm_uri": "ncm_uri_1",
            "node_id": podName,
            "node_name": podName,
            "server_port": 50001,
            #"veth": "eth0"
            "veth": podName.replace("-", "") + "-eth1"
        }

        nodes_body["host_infos"].append(node_body)

    print(nodes_body)

    address = "10.213.43.224"
    nmm_port = "30007"
    headers = {'Content-Type': 'application/json'}
    create_node_endpoint = "http://{}:{}/nodes/bulk".format(address, nmm_port)

    response = requests.post(create_node_endpoint,
                             headers=headers,
                             verify=False,
                             data=json.dumps(nodes_body))
    print(response.text)
    while not response.ok:
        sleep(5)
        # syslog("create_node response {}".format(response.text))
        response = requests.get(create_node_endpoint)
        print(response.text)

main()