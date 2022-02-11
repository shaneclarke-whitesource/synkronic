# Synkronic 
Rainbow Deployment for Kubernetes 


## Design

### Creating Multiple Deployments
A developer will first make normal Deployment & Service, with a specific image.  
Rainbow Alpha Operator [generate service/deploy] can then be configured with new versions referring to that Deployment/Service, and will create a clone of the base Deploy/Service for each new version.  The operator will remove old versions by API call or through reconciliation. 

### Routing
NGinx or Traefik are capable of mapping patterns of subdomain names (version-1.mydomain.com) or URL paths (mydomain.com/version-1) to services.  This may be handled by naming convention between the Cyan Alpha Operator and the Reverse Proxy.  

#### Sample CRD
```
apiVersion: "stable.example.com/v1"
kind: CronTab
metadata:
  name: my-new-cron-object
spec:
  cronSpec: "* * * * */5"
  image: my-awesome-cron-image
```

#### Target CRD
```
apiVersion: "kyaninus.com/v1"
kind: DeploymentVersion
metadata:
  name: my-app-deploy-version-1
  namespace: my-app-namespace
spec:
  name: my-app-deploy-version-1
  namespace: my-app-namespace
  deploymentSpec:
replicas: 1
template:
   spec:
      securityContext:
        	runAsNonRoot: true
     		containers:
      	   name: manager
         image: controller:latest
```

## Development Getting Starting  

### Golang Install
https://golang.org/doc/install

### Minikube Install 
https://v1-18.docs.kubernetes.io/docs/tasks/tools/install-minikube/

### Minikube Docker Repository Setup

https://gist.github.com/trisberg/37c97b6cc53def9a3e38be6143786589

Setup docker registry outside minikube

docker run -d -p 5000:5000 --restart=always --volume ~/.registry/storage:/var/lib/registry registry:2

Set dev host and minikube docker daemons to allow that insecure registry 

Edit the /etc/hosts file on your development machine, adding the name mkregistry.local on the same line as the entry for localhost.
Set host aliases to the registry on host, and inside minikube
Edit /etc/docker/daemon.json (create the file if it does not exist)
```
{
  "insecure-registries": ["mkregistry.local:5000"]
}
```

Configure a fixed IP address

This IP address will allow processes in Minikube to reach the registry running on your host. Configuring a fixed IP address avoids the problem of the IP address changing whenever you connect your machine to a different network. If your machine already uses the 172.16.x.x range for other purposes, choose an address in a different range e.g. 172.31.x.x..

```
export DEV_IP=172.16.1.1
sudo ifconfig lo:0 $DEV_IP
```

Note that the alias will need to be reestablished when you restart your machine. This can be avoided by editing /etc/network/interfaces on Linux.
 
Minikube /etc/hosts

Add an entry to /etc/hosts inside the minikube VM, pointing the registry to the IP address of the host. This will result in registry.dev.svc.cluster.local resolving to the host machine allowing the docker daemon in minikube to pull images from the local registry. This uses the DEV_IP environment variable from the previous step.

```
export DEV_IP=172.16.1.1
minikube ssh "echo \"$DEV_IP       mkregistry.local.local\" | sudo tee -a  /etc/hosts"
```

#### Restarting Minikube
minikube start --insecure-registry mkregistry.local:5000

export DEV_IP=172.16.1.1
minikube ssh "echo \"$DEV_IP       mkregistry.local\" | sudo tee -a  /etc/hosts"

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


