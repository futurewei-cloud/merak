#!/bin/bash

IFS=$'\n'
nodeList=$(kubectl get nodes -o wide | awk '{print $6}')

#echo $gwList
for hostIP in $nodeList
do
  
    echo $hostIP
    ssh -i alcor-distrinet-test.pem ubuntu@$hostIP "sudo rm /etc/cni/net.d/00-meshnet.conflist"
    ssh -i alcor-distrinet-test.pem ubuntu@$hostIP "sudo rm /etc/cni/net.d/10-flannel.conflist"

    echo "$hostIP net.d cleaned"

done



