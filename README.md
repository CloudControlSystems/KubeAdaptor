# KubeAdaptor: A Docking Framework for Workflow Containerization on Kubernetes
We have open-sourced the KubeAdaptor, which can be incorporated into workflow systems for K8s as a docking framework.
You can deploy KubeAdaptor in one-button mode with just a few tweaks to your systematic configuration file. 
We welcome you to download, learn, and work together to maintain the KubeAdaptor with us. If you use it for scientific research and 
engineering applications, please be sure to protect the copyright and indicate authors and source.

##Resource description

###./experiment

This directory includes Argo, Batch Job, and KubeAdaptor workflow submission deployment files.
The 'no_ssh' directory takes care of unlocking ssh password between the cluster Master and 
the cluster nodes.
Three submission approaches contain resource gathering module, respectively.
You can get the 'usage.txt' , 'log.txt' and 'exp.txt' files through our designed automatic scripts.
The automatic script file has the abilities of capturing the Master IP and obtaining the other nodes' IP.
Our experimental environment consists of one Master and six nodes. In K8s cluster, node name is identified by node's IP.

OS version: Ubuntu 20.4/CentOS Linux release 7.8, Kubernetes: v1.18.6/v1.19.6, Docker: 18.09.6.

Note that all the image addresses in the following yaml file need to remove the string 
'harbor.cloudcontrolsystems.cn/'. In this way, your image address become the Docker Hub address.

1. ./experiment/argo_test

steps:

##### a. kubectl create ns argo && kubectl apply -f install.yaml -n argo

  The two commands finish the creation of Argo namespace and Argo's installation.

##### b. Select workflow and do a few tweaks for configuration files

Update the number of 'ClusterRoleBinding' of the 'rbac-argo.yaml.bak' in line with the number of cluster nodes, and 
set the field of 'system:node' to '****'. The character '*' represents the node number.
Update the field 'node.num' to the number of cluster nodes. The field 'gather.time' selects 
  the sample cycle to 500 milliseconds in default. 
In 'resourceUsage.yaml.bak', the field 'node.name' and field 'gather.time' are defined as the same as above.
Each workflow definition corresponds to the respective 'workflowName.yaml.bak' file.

Then, you can select one workflow to be tested in 'edit.sh' file, such as the command 'cp ./Montage.yaml.bak Montage.yaml'.

##### c. ./deploy.sh

This shell script file is responsible for deploying RBAC permission and resource gathering module.
After the resource gathering module pod has launched, this script consecutively
submits one workflow for 100 times through argo binary tools. In the end, we can obtain various of log files.

You can list the testing workflow through Argo binary tool.

  argo list -n argo

  argo get -n argo @taskName

  argo logs -n argo @taskName

##### d. ./clean.sh

Clear up RBAC permission and resource gathering module.

2. ./experiment/batch_job

We define four real-world workflows composed of batch yaml files. Take the cybershake workflow as example.
In './experiment/batch_job/cybershake' directory,  the 'deleteLog.sh' take cares of cleaning up 
various log files, wfScheduler.sh is responsible for deploying each batch workflow tasks topologically.
We define the yaml file of each batch workflow tasks through 'PriorityClass' method.
Before running this workflow, you need to update the 'rbac-deploy.yaml' and 'resourceUsage.yaml'.
The field 'system:node' of 'rbac.deploy.yaml' corresponds to each cluster node' IP.
The 'resourceUsage.yaml' file defines the field 'node.num' and field 'gather.time' in accordance with
'resourceUsage.yaml.bak' of Argo submission method.

steps:

##### ./wfScheduler.sh

The shell script is responsible for deploying the RBAC and resource gathering module and running this 
real-world workflow. You can set the for loop as many times as you want. 
In the end, we can obtain various of log files.

3. ./experiment/KubeAdaptor_test

The directory './deploy' includes the Yaml files corresponding to KubeAdaptor, RBAC, resource usage rate, Nfs, and workflow injection module.
We use Configmap method in Yaml file to inject workflow information (dependency.json) into the container of workflow injection module.
Refer to './deploy/buildInjector1.yaml' for details.

steps:

##### a. update the K8s nodes' ip.

Update the 'ipNode.txt' in line with your test cluster.

##### b. ./deploy.sh

Deploy the KubeAdaptor into the K8s cluster. The './deploy/edit.sh' file firstly captures the Master's IP,
updates the other corresponding files. Then it copies '$workflowName.yaml.bak' to 'workflowInjector-Builder.yaml'.
The 'deploy.sh' file includes a series of 'Kubectl' commands.
During the workflow lifecycle, you can watch the execution states of workflow tasks. The following is the operation command.

'kubectl get pods -A --watch -o wide'

##### c. ./clear.sh

When the workflow is completed, you can run '.clear.sh' file to clean up the workflow information and obtain the log files.

4. ./experiment/no_ssh
   
The 'no_ssh' directory takes care of unlocking ssh password between the cluster Master and the cluster nodes.

###./resourceUsage

This directory includes the source codes of resource gathering module.
You can build Docker image by 'Dockerfile' file or pull the image of this module from docker Hub.

'docker pull shanchenggang/resource-usage:v1.0'

###./TaskContainerBuilder

This directory includes the source codes of CWB.
You can build Docker image by 'Dockerfile' file or pull the image of this module from docker Hub.

'docker pull shanchenggang/task-container-builder:v5.0'

###./usage_deploy

This directory includes the deployment file of resource gather module.
Noting that './experiment' directory has included this deployment file for three submission approaches.

###./WorkflowInjector

This directory includes the source codes of the workflow injection module.
You can build Docker image by 'Dockerfile' file or pull the image of this module from docker Hub.

'docker pull shanchenggang/workflow-injector:v5.0'
