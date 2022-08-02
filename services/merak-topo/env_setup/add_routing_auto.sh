#!/bin/bash

IFS=$'\n'
gwList=$(kubectl get pod -A -o wide | grep cgw-[0-9] | awk '{print $7, $8}')

#echo $gwList
for gw in $gwList
do
#       echo gw
        ip=$(echo $gw | cut -d' ' -f1)
    hostIPs=$(echo $gw | cut -d' ' -f2)
    hostIP1=$(echo $hostIPs | cut -d'-' -f2)
    hostIP2=$(echo $hostIPs | cut -d'-' -f3)
    hostIP3=$(echo $hostIPs | cut -d'-' -f4)
    hostIP4=$(echo $hostIPs | cut -d'-' -f5)
    hostIP=$(echo $hostIP1"."$hostIP2"."$hostIP3"."$hostIP4)
    echo $ip " and " $hostIP
#ssh root@172. "route add -net xxx"
#echo "172.xxx route added"

    ssh -i ym-keypair.pem ubuntu@$hostIP "sudo route add -net 10.200.0.0/16 gw $ip"
    echo "$hostIP route added"

done


