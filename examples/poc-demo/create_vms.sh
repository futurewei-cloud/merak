#!/bin/bash

function create_vms() {
   local pod=$1
   local subnets=$2

   if [ ! -d ~/vm_logs ]; then
        mkdir -p ~/vm_logs
   fi

   echo "*** create vm in $pod with subnet $subnets" 2>&1 | tee -a ~/vm_logs/$pod.log
   kubectl cp interface_create.py $pod:. 2>&1 | tee -a ~/vm_logs/$pod.log
   kubectl exec -i $pod -- pip3 install requests 2>&1 | tee -a ~/vm_logs/$pod.log
   echo "*** command: kubectl exec -i $pod -- python3 interface_create.py -s $subnets" 2>&1 | tee -a ~/vm_logs/$pod.log
   kubectl exec -i $pod -- python3 interface_create.py -s $subnets 2>&1 | tee -a ~/vm_logs/$pod.log

}


if [[ $# -lt 2 ]]; then
        echo "Usage: create_vms <number of vms>  <subnet-id>"
        exit 1
else
        NUMOFVMS=$1
        SUBNET=$2
fi

for i in $(eval echo "{1..$NUMOFVMS}")
do
   SUBNETS=$SUBNET" "$SUBNETS   
done	

IFS=$'\n'
podList=$(kubectl get pods -A | grep vhost-[0-9] | awk '{print $2}')

#echo $gwList
for podname in $podList
do
   echo "$podname start to create VMs"
   create_vms $podname $SUBNETS 
   echo "$podname create VM done!"
   sleep 2
done
