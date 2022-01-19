# KubeAdaptor: A Docking Framework for Workflow Containerization on Kubernetes
We have open-sourced the KubeAdaptor, which can be incorporated into workflow systems for K8s as a docking framework.
You can deploy KubeAdaptor in one-button mode with just a few tweaks to your systematic configuration file. 
We welcome you to download, learn, and work together to maintain the KubeAdaptor with us. If you use it for scientific research and 
engineering applications, please be sure to protect the copyright and indicate authors and source.

## Resource description

### ./experiment

This directory includes Argo, Batch Job, and KubeAdaptor workflow submission deployment files.
The `no_ssh` directory takes care of unlocking ssh passwords between the cluster Master and 
the cluster nodes.
Three submission approaches contain resource gathering modules, respectively.
You can get the `usage.txt`, `log.txt`, and `exp.txt` files through our designed automatic scripts.
The automatic script file has the ability to capture the Master IP and obtain the other nodes' IP.
Our experimental environment consists of one Master and six nodes. In the K8s cluster, the node name 
is identified by the node's IP.

#### Software Prerequisites

1. OS version: Ubuntu 20.4/CentOS Linux release 7.8
2. Kubernetes: v1.18.6/v1.19.6
3. Docker: 18.09.6.

Note that all the image addresses in the following YAML file need to remove the string 
`harbor.cloudcontrolsystems.cn`. In this way, your image address becomes the `Docker Hub address`.
In addition, you need to deploy the `NFS server service` into your cluster in advance so that each node is able to
 mount the Master node's shared directory.
If you have any questions during the operation, please do not hesitate to contact `uzz_scg@163.com`.

#### ./experiment/argo_test

**steps:**

* Run `kubectl create ns argo && kubectl apply -f install.yaml -n argo`

  The two commands finish the creation of Argo namespace and Argo's installation.

* Select workflow and do a few tweaks for configuration files

  Update the number of `ClusterRoleBinding` field of the `rbac-argo.yaml.bak` in line with the number of cluster nodes, and 
set the field of `system:node` to `****`. The character `\*` represents the node number.
Update the field `node.num` to the number of cluster nodes. The field `gather.time` selects 
  the sample cycle to 500 milliseconds in default. 
In `resourceUsage.yaml.bak`, the field `node.name` and field `gather.time` are defined as the same as above.
Each workflow definition corresponds to the respective `$workflowName.yaml.bak` file.

  Then, you can select one workflow to be tested in `edit.sh` file, such as the command `cp ./Montage.yaml.bak Montage.yaml`.

* Run `./deploy.sh`

  This shell script file is responsible for deploying the RBAC permission and resource gathering module.
After the resource gathering module pod has launched, this script consecutively
submits one workflow 100 times through Argo binary tools. In the end, we can obtain various log files.
  
  You can list the testing workflow through the Argo binary tool.
```console
  argo list -n argo
  argo get -n argo @taskName
  argo logs -n argo @taskName
```
* Run `./clean.sh`

  Clear up the RBAC permission and resource gathering module.

#### ./experiment/batch_job

We define four real-world workflows composed of batch YAML files. Take the cybershake workflow as an example.
In the `./experiment/batch_job/cybershake` directory,  the `deleteLog.sh` takes care of cleaning up 
various log files, the `wfScheduler.sh` is responsible for deploying each batch workflow task topologically.
We define the YAML file of a batch of workflow tasks through the `PriorityClass` method.
Before running this workflow, you need to update the `rbac-deploy.yaml` and `resourceUsage.yaml`.
The field `system:node` of `rbac.deploy.yaml` corresponds to the respective node' IP.
The `resourceUsage.yaml` file defines the field `node.num` and field `gather.time` in line with
the `resourceUsage.yaml.bak` file of the Argo submission method.

**steps:**

* Run `./wfScheduler.sh`

  The shell script is responsible for deploying the RBAC and resource gathering module and 
running this real-world workflow. 
You can set the for loop as many times as you want. In the end, 
this script automatically cleans up each module and obtains various log files.

#### ./experiment/KubeAdaptor_test

The directory `./deploy` includes the Yaml files corresponding to KubeAdaptor, RBAC, resource usage rate, Nfs, and workflow injection module.
We use the Configmap method in Yaml file to inject workflow information (dependency.json) into the container of the workflow injection module.
Refer to `./deploy/buildInjector1.yaml` for details.

**steps:**

* Update the K8s nodes' ip.

  Update the `ipNode.txt` in line with your test cluster.

* Run `./deploy.sh`

  Deploy the KubeAdaptor into the K8s cluster. The `./deploy/edit.sh` file firstly captures the Master's IP,
updates the other corresponding files. Then it copies `$workflowName.yaml.bak` to `workflowInjector-Builder.yaml`.
The `deploy.sh` file includes a series of `Kubectl` commands.
During the workflow lifecycle, you can watch the execution states of workflow tasks. The following is the operation command.
  ```console
  'kubectl get pods -A --watch -o wide'
  ```
* Run `./clear.sh`

  When the workflow is completed, you can run the `./clear.sh` file to clean up the workflow information and obtain the log files.

### ./experiment/no_ssh
   
  The `no_ssh` directory takes care of unlocking ssh passwords between the cluster Master and the cluster nodes.

### ./resourceUsage

This directory includes the source codes of the resource gathering module.
You can build the Docker image by the `Dockerfile` file or pull the image of this module from the Docker Hub.
```console
docker pull shanchenggang/resource-usage:v1.0
```
### ./TaskContainerBuilder

This directory includes the source codes of KubeAdaptor.
You can build the Docker image by the `Dockerfile` file or pull the image of this module from Docker Hub.
```console
docker pull shanchenggang/task-container-builder:v6.0
```
### ./usage_deploy

This directory includes the deployment file of the resource gather module.
Note that the `./experiment` directory has included this deployment file for three submission approaches.

### ./WorkflowInjector

This directory includes the source codes of the workflow injection module.
You can build Docker image by `Dockerfile` file or pull the image of this module from Docker Hub.
```console
docker pull shanchenggang/workflow-injector:v6.0
```