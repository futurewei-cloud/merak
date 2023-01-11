#!/bin/bash
yes | sudo kubeadm reset
sudo rm -rf /root/work && sudo kubeadm init --pod-network-cidr 10.244.0.0/16
mkdir -p $HOME/.kube
yes | sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
kubectl taint nodes --all node-role.kubernetes.io/control-plane:NoSchedule-
kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml
kubectl kustomize deployments/kubernetes/compute.test --enable-helm | kubectl apply -f -