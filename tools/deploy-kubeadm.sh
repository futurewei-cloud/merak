#!/bin/bash
yes | sudo kubeadm reset
sudo rm -rf /root/work && sudo kubeadm init --pod-network-cidr 10.244.0.0/16
mkdir -p $HOME/.kube
yes | sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
kubectl taint nodes --all node-role.kubernetes.io/control-plane:NoSchedule-
kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml
linkerd install --crds | kubectl apply -f -
linkerd install --set proxyInit.runAsRoot=true | kubectl apply -f -
linkerd check
kubectl kustomize deployments/kubernetes/test --enable-helm | kubectl apply -f -
