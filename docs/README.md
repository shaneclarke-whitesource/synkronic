# Synkronic 
## Rainbow Deployment for Kubernetes 

Rainbow deployment addresses congestion that occurs in development processes, where teams must sync and merge codebases, stabilize branches, coordinate releases all while managing workstreams, breakages and other exceptional cases.  In agile processes an unstable feature branch will often merge and deploy to environment, delaying or forcing a repeat of the release process.  

Rainbow deployment allows for an unlimited number of versions of an application to be available, and so tested and ensured to be stable before merging into any other branch.  This enables change reviewers to view development work without waiting for a coordinated release, but instead as soon as the developer is ready to push.  The flexibility of these feature branch reviews extend into higher environments, where multiple versions of a production service may run and enable blue/green style testing before activating a new version, and with versions remaining for an extended time greater options in rolling back.  Further, rainbow deployments allow for simultaneous testing of multiple versions by QA, or partial releases to subsets of service consumers.  A QA team may be targeted to test one version while another team can test a second, both receiving frequent updates.  

## Design

### Multiple Version Spin-up
A developer will first deply a service by making normal Deployment & Service objects.

A Synkronic Operator will run in the environment, configured by a CRD object which specifies parameters for new versions of the service.  As new versions are configured using operator CRDs, the base objects will be cloned and created using these new parameters.  Removal or modification of these CRDs will cause removal of the new versions.  

This CRD is a *DeploymentVersion* object.  It contains references to the base resources it will cause to clone, and the parameters described above.  

Another CRD *Router* may configure how the operator will modify Ingress or othre services in response to creation of a new version.  

### Routing to Versions
NGinx or Traefik are capable of mapping patterns of subdomain names (version-1.mydomain.com) or URL paths (mydomain.com/version-1) to services.  This may be handled by naming convention between the Synkronic Operator and the Reverse Proxy.  


#### Sample CRD
This is a sample custom resource definition.  

```yaml
apiVersion: "stable.example.com/v1"
kind: CronTab
metadata:
  name: my-new-cron-object
spec:
  cronSpec: "* * * * */5"
  image: my-awesome-cron-image
```

#### DeploymentVersion CRD
```yaml
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
          containers:
          - name: manager
            image: controller:latest
```



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


