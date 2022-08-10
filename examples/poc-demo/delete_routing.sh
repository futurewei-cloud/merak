#!/bin/bash

IFS=$'\n'
gwList=$(kubectl get pods -A -o wide | grep cgw-[0-9] | awk '{print $7, $8}')

#echo $gwList
for gw in $gwList
do
    ip=$(echo $gw | cut -d' ' -f1)
    hostName=$(echo $gw | cut -d' ' -f2)
    hostIP=$(ping -c1 $hostName | sed -nE 's/^PING[^(]+\(([^)]+)\).*/\1/p')
    ssh root@$hostIP "sudo route del -net 10.200.0.0/16 gw $ip"
    echo "'sudo route del -net 10.200.0.0/16 gw $ip' deleted from $hostIP"
done
