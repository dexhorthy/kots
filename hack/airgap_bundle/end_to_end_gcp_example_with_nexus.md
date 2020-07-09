### Appendix: End to End GCP example

We'll set up these 3 instances in GCP. Unless otherwise specified, all commands are being run from a MacOS workstation outside this environment.

```
airgap-jump                                      us-central1-b  n1-standard-1                10.240.0.127  35.193.94.81     RUNNING
airgap-cluster                                   us-central1-b  n1-standard-1                10.240.0.41                    RUNNING
airgap-workstation                               us-central1-b  n1-standard-1                10.240.0.26                    RUNNING
```


**Note**: This guide does a lot of network configuration for address management, but omits any details regarding opening ports. While you could open specific ports between instances, this guide was written with inter-instance traffic wide open.

We'll use ssh tunneling for reaching the instances in the cluster, so it shouldn't be necessary to open ports for access from the outside world.

#### jump box

Create an jump box with a public IP and SSH it, this will be our jump box w/ internet access and also access to the airgapped environment


```shell script
export INSTANCE=airgap-jump; gcloud compute instances create $INSTANCE --boot-disk-size=200GB --image-project ubuntu-os-cloud --image-family ubuntu-1804-lts --machine-type n1-standard-1
```

#### airgapped workstation

create a GCP vm to be our airgapped workstation. We'll give it outbound network access for now to facilitate installing docker, but then we'll disconnect it from the internet. Replace `dex` in the `usermod` command with your unix username in GCP.


```shell script
export INSTANCE=airgap-workstation; gcloud compute instances create $INSTANCE --boot-disk-size=200GB --image-project ubuntu-os-cloud --image-family ubuntu-1804-lts --machine-type n1-standard-1 
```

```shell script
export LINUX_USER=dex
gcloud compute ssh airgap-workstation -- 'sudo apt update && sudo apt install -y docker.io'
gcloud compute ssh airgap-workstation -- "sudo usermod -aG docker ${LINUX_USER}"
gcloud compute ssh airgap-workstation -- 'sudo snap install kubectl --classic'
```

Let's also pull a standard busybox image before turning off internet, we'll use this for testing later

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- ssh airgap-workstation -- docker pull busybox
```

Next, remove the machine's public IP. 

```shell script
gcloud compute instances delete-access-config airgap-workstation
```

verify that internet access was disabled by ssh'ing via the jump box and trying to curl kubernetes.io. We'll forward the agent so that we can ssh the airgapped workstation without moving keys around

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- "ssh airgap-workstation 'curl -v https://kubernetes.io'"
```

this command should hang, and you should see something with `Network is unreachable`:

```text
  0     0    0     0    0     0      0      0 --:--:--  0:00:02 --:--:--     0*   Trying 2607:f8b0:4001:c05::64...
* TCP_NODELAY set
* Immediate connect fail for 2607:f8b0:4001:c05::64: Network is unreachable
  0     0    0     0    0     0      0      0 --:--:--  0:00:03 --:--:--     0
```


#### Registry Access

This guide assumes we have an existing docker registry, in this case we'll be using a Nexus3 OSS instance w/ a docker registry configured

Let's first verify our credentials from the airgapped workstation and check that we can push/pull an image. **NOTE** in this case, the registry is deployed in the same private network as our airgap-workstation, and we'll access it via private IP, but the only requirement is that the Nexus instance be reachable from the airgapped workstation

```shell script
export DOCKER_PASSWORD=...
export DOCKER_USERNAME=...
export DOCKER_REGISTRY=10.0.0.127
gcloud compute ssh --ssh-flag=-A airgap-jump -- ssh airgap-workstation -- docker login --username ${DOCKER_USERNAME} --password ${DOCKER_PASSWORD} ${DOCKER_REGISTRY}
gcloud compute ssh --ssh-flag=-A airgap-jump -- ssh airgap-workstation -- docker tag busybox ${DOCKER_REGISTRY}/busybox
gcloud compute ssh --ssh-flag=-A airgap-jump -- ssh airgap-workstation -- docker push ${DOCKER_REGISTRY}/busybox
```

#### airgapped cluster 

create a GCP vm with online internet access, this will be our airgapped cluster, but we'll use a an internet connection to install k8s and get a registry up and running.

```shell script
INSTANCE=airgap-cluster; gcloud compute instances create $INSTANCE --boot-disk-size=200GB --image-project ubuntu-os-cloud --image-family ubuntu-1804-lts --machine-type n1-standard-4 
```


Now, let's ssh into the instance and bootstrap a minimal kubernetes cluster (details here:  https://kurl.sh/1010f0a  )

```shell script
gcloud compute ssh airgap-cluster -- 'curl  https://k8s.kurl.sh/1010f0a  | sudo bash'
```

Next, remove the machine's public IP. We'll use the kubeconfig from this server later.

```shell script
gcloud compute instances delete-access-config airgap-cluster
```

verify that internet access was disabled by ssh'ing via the jump box and trying to curl kubernetes.io. We'll forward the agent so that we can ssh the airgapped cluster without moving keys around

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- "ssh airgap-cluster 'curl -v https://kubernetes.io'"
```

this command should hang, and you should see something with `Network is unreachable`:

```text
  0     0    0     0    0     0      0      0 --:--:--  0:00:02 --:--:--     0*   Trying 2607:f8b0:4001:c05::64...
* TCP_NODELAY set
* Immediate connect fail for 2607:f8b0:4001:c05::64: Network is unreachable
  0     0    0     0    0     0      0      0 --:--:--  0:00:03 --:--:--     0
```


#### Final Workstation Setup


Now, let's very our docker client on the workstation and make sure we have kubectl access properly configured before we do the full installation. We'll do by ssh'ing the workstation via the jump box


###### Kubectl

next, ssh into the airgapped worksation and grab the `admin.conf` from the cluster and run a few kubectl commands to ensure its working

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- 'ssh -A airgap-workstation'
```

From the Airgapped workstation, run the following:

```shell script
scp airgap-cluster:admin.conf .
export KUBECONFIG=$PWD/admin.conf
kubectl get ns
kubectl get pod -n kube-system
```

Should see something like

```
NAME                                   READY   STATUS    RESTARTS   AGE
coredns-5644d7b6d9-j6gqs               1/1     Running   0          15m
coredns-5644d7b6d9-s7q64               1/1     Running   0          15m
etcd-dex-airgap-2                      1/1     Running   0          14m
kube-apiserver-dex-airgap-2            1/1     Running   0          14m
kube-controller-manager-dex-airgap-2   1/1     Running   0          13m
kube-proxy-l6fw8                       1/1     Running   0          15m
kube-scheduler-dex-airgap-2            1/1     Running   0          13m
weave-net-7nf4z                        2/2     Running   0          15m
```

Now -- log out of the airgapped instance

```shell script
exit
```

###### Namespace and Secret

One of the prerequisites for the installer is a namespace with an existing pull secret for the install, let's create those now:


```shell script
export NAMESPACE=test-deploy
gcloud compute ssh --ssh-flag=-A airgap-jump -- "ssh -A airgap-workstation -- /snap/bin/kubectl --kubeconfig=admin.conf create namespace ${NAMESPACE}"
```

Should show

```text
namespace/test-deploy created
```

Next, let's make a secret for our registry

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- "ssh -A airgap-workstation -- /snap/bin/kubectl --kubeconfig=admin.conf -n $NAMESPACE create secret  docker-registry registry-creds --docker-server=${DOCKER_REGISTRY} --docker-username=${DOCKER_USERNAME} --docker-password=${DOCKER_PASSWORD} --docker-email=a@b.c"
```

We should see

```text
secret/registry-creds created
```

#### Installing

From the Jump box, download the kots bundle from S3 and scp it to the airgapped workstation. In a "full airgap" or "sneakernet" scenario, replace `scp` with whatever process is appropriate for moving assets into the airgapped cluster.

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- 'wget https://kots-experimental.s3.amazonaws.com/kots-v1.16.2-airgap-experimental-alpha4.tar.gz'
gcloud compute ssh --ssh-flag=-A airgap-jump -- 'scp kots-v1.16.2-airgap-experimental-alpha4.tar.gz airgap-workstation:'
```

Now, we're ready to untar the bundle and run the install script:


```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- 'ssh airgap-workstation tar xvf kots-v1.16.2-airgap-experimental-alpha4.tar.gz'
```


Should print

```text

./
./support-bundle
./troubleshoot/
./LICENSE
./images/
./install.sh
./README.md
./kots
./yaml/
./yaml/kotsadm.yaml
./images/kotsadm-kotsadm-migrations-v1.16.2.tar
./images/postgres-10.7.tar
./images/kotsadm-kotsadm-operator-v1.16.2.tar
./images/kotsadm-kotsadm-api-v1.16.2.tar
./images/kotsadm-kotsadm-v1.16.2.tar
./images/kotsadm-minio-v1.16.2.tar
./troubleshoot/support-bundle.yaml
```

Next, let's run it with our parameters, passing the registry IP, namespace we created, and name of the registry secret. We'll omit the `DOCKER_USERNAME` and `DOCKER_PASSWORD` args since we've previously run a `docker login` on the workstation

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- "ssh airgap-workstation -- KUBECONFIG=./admin.conf PATH=${PATH}:/snap/bin ./install.sh ${DOCKER_REGISTRY} ${NAMESPACE} registry-creds "
```


Full output should look something like


```text


==========
Checking for prerequisites: namespace exists, with image pull secret, and registry push credentials are valid
==========
SUCCESS: namespace "test-deploy" exists
SUCCESS: secret "registry-creds" exists
SUCCESS: no push credentials provided, skipping docker login.
Error from server (NotFound): deployments.apps "kotsadm-api" not found
SUCCESS: it appears that namespace "test-deploy" does not contain any existing KOTS resources from a previous deploy

==========
preparing kustomization
==========
    MIGRATIONS_POD_NAME=kotsadm-migrations-1593530428
    AUTO_CREATE_CLUSTER_TOKEN=WnMVWPwhrabyAmbu
SUCCESS: valid kustomization yaml created in ./yaml

==========
loading docker images
==========
Loaded image: kotsadm/kotsadm-api:v1.16.1
Loaded image: kotsadm/kotsadm-migrations:v1.16.1
Loaded image: kotsadm/kotsadm-operator:v1.16.1
Loaded image: kotsadm/kotsadm:v1.16.1
Loaded image: kotsadm/minio:v1.16.1
Loaded image: postgres:10.7

==========
tagging and pushing to 10.240.0.88:32000
==========
The push refers to repository [10.240.0.88:32000/kotsadm-api]
9906809c4536: Preparing
64e44f6ee017: Preparing
64fb7723a8c9: Preparing
...
etc etc etc
...


==========
deploying
==========
manifests have been written to ./yaml -- you can press ENTER to deploy them, or Ctrl+C to exit this script. You can deploy them later with

    kubectl apply --namespace test-deploy -k ./yaml

would you like to deploy? [ENTER] 
serviceaccount/kotsadm-api created
serviceaccount/kotsadm-operator created
serviceaccount/kotsadm created
role.rbac.authorization.k8s.io/kotsadm-api-role created
role.rbac.authorization.k8s.io/kotsadm-operator-role created
clusterrole.rbac.authorization.k8s.io/kotsadm-role created
rolebinding.rbac.authorization.k8s.io/kotsadm-api-rolebinding created
rolebinding.rbac.authorization.k8s.io/kotsadm-operator-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/kotsadm-rolebinding created
secret/kotsadm-cluster-token created
secret/kotsadm-encryption created
secret/kotsadm-minio created
secret/kotsadm-password created
secret/kotsadm-postgres created
secret/kotsadm-session created
service/kotsadm-api-node created
service/kotsadm-minio created
service/kotsadm-postgres created
service/kotsadm created
deployment.apps/kotsadm-api created
deployment.apps/kotsadm-operator created
deployment.apps/kotsadm created
statefulset.apps/kotsadm-minio created
statefulset.apps/kotsadm-postgres created
pod/kotsadm-migrations-1593530428 created

==========
postflight checks
==========
SUCCESS: cluster token configured
Connection to 34.68.172.116 closed.
```

check the pods and wait for things to come up:

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- ssh airgap-workstation -- KUBECONFIG=./admin.conf /snap/bin/kubectl -n "${NAMESPACE}" get pod
```


### Connecting to KOTS

Now that we're installed, we need to connect in. We'll use a NodePort and an ssh tunnel, but based on your cluster you could also create an ingress or access via a kubectl port-forward if you have access from a workstation.

First though, we'll reset the password:

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- ssh airgap-workstation -- KUBECONFIG=./admin.conf ./kots reset-password -n "${NAMESPACE}"
```

Enter any password you like:

```text
  • Reset the admin console password for test-deploy
Enter a new password to be used for the Admin Console: █
Enter a new password to be used for the Admin Console: ••••••••
  • The admin console password has been reset
```

Now, we'll create a node port to expose the service


```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- ssh airgap-workstation -- KUBECONFIG=./admin.conf /snap/bin/kubectl -n "${NAMESPACE}" expose deployment kotsadm --name=kotsadm-nodeport --port=3000 --target-port=3000 --type=NodePort
```

Next, we need to get the port and expose it locally via an SSH tunnel

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- ssh airgap-workstation -- KUBECONFIG=./admin.conf /snap/bin/kubectl -n "${NAMESPACE}" get svc kotsadm-nodeport
```

Asumming this is our output, we'll set the `PORT` to `40038`

```shell script
NAME               TYPE       CLUSTER-IP   EXTERNAL-IP   PORT(S)          AGE
kotsadm-nodeport   NodePort   10.96.3.54   <none>        3000:40038/TCP   6s
```


Create a SSH tunnel on your laptop via the Jumpbox node.

```shell script
export CLUSTER_PRIVATE_IP=$(gcloud compute instances describe airgap-cluster --format='get(networkInterfaces[0].networkIP)')
export PORT=40038
gcloud compute ssh --ssh-flag=-N --ssh-flag="-L ${PORT}:${CLUSTER_PRIVATE_IP}:${PORT}" airgap-jump
```


Now, open `localhost:${PORT}` in your browser and you should get to the kotsadm console, proceeding with the install from there.

### Troubleshooting

If you run into issues, you may be able to use the bundled support-bundle tool to collect a very helpful diagnostic bundle. This will only be usable once 

- the cluster is up and 
- you have the `admin.conf` kubeconfig on the airgap workstation
- you have unpacked the kots tar.gz on the airgap workstation

The support bundle collected will include logs for all kots services:

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- ssh airgap-workstation -- KUBECONFIG=./admin.conf ./support-bundle ./troubleshoot/support-bundle.yaml
```

then, copy the bundle to your local machine

```shell script
gcloud compute ssh --ssh-flag=-A airgap-jump -- scp airgap-workstation:support-bundle.tar.gz .
gcloud compute scp airgap-jump:support-bundle.tar.gz .
```

### Cleaning up

To clean up, delete the servers in question

```shell script
gcloud compute instances delete airgap-cluster airgap-jump airgap-workstation
```
