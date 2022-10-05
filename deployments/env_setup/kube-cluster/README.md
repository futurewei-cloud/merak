# This doc is for set up the k8s env  
Suggest to put this on the k8s control node.
### The scripts can:
1. Install k8s
2. Pull docker images
3. Install dependencies on each compute node  
4. Remove `/root/work/` folder on each compute node  

### First need to install ansible in venv:
___

```
sudo apt update
sudo apt upgrade
sudo apt install python3-dev python3-venv libffi-dev gcc libssl-dev git

python3 -m venv ansible
source ansible/bin/activate

pip install -U pip
pip install 'ansible<2.10'
```

### Set up k8s
___  

Run:  
`ansible-playbook -i aws-host kube-dependencies.yml`  

After the above command, run the follow to init k8s on the master node:  
`kubeadm init --pod-network-cidr=10.244.0.0/16` 

Then run and put the follow line in `.profile` to be able to manage k8s:  
`export KUBECONFIG=/etc/kubernetes/admin.conf`  

To install flannel:  
`kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml`  

All done with k8s setup!  

## Install other dependencies like ovs:  
___  

`ansible-playbook -i aws-host prepare_machine.yml`  
The current `prepare_machine.yml` file will also install python and python-docker, which then used for pulling docker images later.   

## Pull docker images:  
___  

`ansible-playbook -i aws-host pull_image.yml`  

## Install Linkerd:  
___

To install Linkerd, can following the link or commands below:  

https://linkerd.io/2.12/getting-started/

- Install the Linkerd CLI:  
  - `curl --proto '=https' --tlsv1.2 -sSfL https://run.linkerd.io/install | sh`  
- Export the Linkerd path:  
  - `export PATH=$PATH:/home/ubuntu/.linkerd2/bin`  
-  Validate Kubernetes cluster:
  - `linkerd check --pre`  
- Install Linkerd:  
  - `linkerd install --crds | kubectl apply -f -`  
  - `linkerd install | kubectl apply -f -`  
- Check if Linkerd is installed properly:  
  - `linkerd check`  




