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
    # list all gw

    cmd = "kubectl get pods"
    returncode, text = run_cmd(cmd)
    # print(returncode)
    # print(text)
    podNames = []
    lines=text.splitlines()

    for line in lines:
        if "cgw-" in line:
            # print(line)
            podName = line.split()[0]
            podNames.append(podName)
    print(podNames)


    for podName in podNames:

        pod_num = podName.split("-")[-1].strip()
        cmd_list=[]
        cmd_list.append("kubectl exec -it " + podName + " -- /bin/bash -c \"iptables -F -t nat\"")
        cmd_list.append("kubectl exec -it " + podName + " -- /bin/bash -c \"iptables -t nat -A POSTROUTING -o " + "cgw" + pod_num +"-eth1 -j MASQUERADE\"")
        cmd_list.append("kubectl exec -it " + podName + " -- /bin/bash -c \"iptables -A FORWARD -i eth0 -o " + "cgw" + pod_num +"-eth1 -m state --state RELATED,ESTABLISHED -j ACCEPT\"")
        cmd_list.append("kubectl exec -it " + podName + " -- /bin/bash -c \"iptables -A FORWARD -i " + "cgw" + pod_num +"-eth1 -o eth0 -m state --state RELATED,ESTABLISHED -j ACCEPT\"")

        print ("####start to set up iptables for  " + podName + "#####")
        flag= False
        for cmd in cmd_list:
            returncode, text = run_cmd(cmd)
            #print("returncode",returncode)
            print("cmd", cmd)
            #print("text",text)
            if returncode != None:
                flag = True

        if flag == True:
            print("### Configurateion Failed###")
        else:
            print("### Configuration Done###")


main()
