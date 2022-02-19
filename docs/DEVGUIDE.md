## Development Getting Starting  

### Golang Install
https://golang.org/doc/install

### Minikube Install 
https://v1-18.docs.kubernetes.io/docs/tasks/tools/install-minikube/

### Minikube Docker Repository Setup

https://gist.github.com/trisberg/37c97b6cc53def9a3e38be6143786589

#### Setup docker registry outside minikube

The registry will host locally built images of the operator.
```bash
docker run -d -p 5000:5000 --restart=always --volume ~/.registry/storage:/var/lib/registry registry:2
```

##### Set dev host and minikube docker daemons to allow that insecure registry 

Edit the /etc/hosts file on your development machine, adding the name mkregistry.local on the same line as the entry for localhost.
Set host aliases to the registry on host, and inside minikube
Edit /etc/docker/daemon.json (create the file if it does not exist)
```json
{
  "insecure-registries": ["mkregistry.local:5000"]
}
```

##### Configure a fixed IP address

This IP address will allow processes in Minikube to reach the registry running on your host. Configuring a fixed IP address avoids the problem of the IP address changing whenever you connect your machine to a different network. If your machine already uses the 172.16.x.x range for other purposes, choose an address in a different range e.g. 172.31.x.x..

```bash
export DEV_IP=172.16.1.1
sudo ifconfig lo:0 $DEV_IP
```

Note that the alias will need to be reestablished when you restart your machine. This can be avoided by editing /etc/network/interfaces on Linux.
 
##### Minikube /etc/hosts

Add an entry to /etc/hosts inside the minikube VM, pointing the registry to the IP address of the host. This will result in registry.dev.svc.cluster.local resolving to the host machine allowing the docker daemon in minikube to pull images from the local registry. This uses the DEV_IP environment variable from the previous step.

```bash
export DEV_IP=172.16.1.1
minikube ssh "echo \"$DEV_IP       mkregistry.local.local\" | sudo tee -a  /etc/hosts"
```

#### Restarting Minikube
```bash
minikube start --insecure-registry mkregistry.local:5000

export DEV_IP=172.16.1.1
minikube ssh "echo \"$DEV_IP       mkregistry.local\" | sudo tee -a  /etc/hosts"
```

#### Operator SDK Installs
https://sdk.operatorframework.io/docs/installation/
https://github.com/kubernetes-sigs/kubebuilder

#### Operator SDK Build & Deploy (to Minikube)
https://book.kubebuilder.io/quick-start.html#create-a-project

make docker-build
make docker-deploy
make install
make deploy

#### Test Setup
https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/book/src/reference/envtest.md
https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html


