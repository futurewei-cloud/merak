#!/bin/bash
export KUBECONFIG=/etc/kubernetes/admin.conf

rm -rf /root/work

for ip_addr in $(cat /root/alcor-nodes-ips); do
	ssh root@$ip_addr "rm -rf /root/work"
	echo $ip_addr deleted
done
