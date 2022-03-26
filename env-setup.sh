export DEV_IP=172.16.1.1
sudo ifconfig lo:0 $DEV_IP
minikube delete
minikube start --insecure-registry mkregistry.local:5000
minikube ssh "echo \"$DEV_IP       mkregistry.local\" | sudo tee -a  /etc/hosts"
