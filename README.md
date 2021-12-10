# Containerized workflow builder for Kubernetes
We have open-sourced the CWB, which can be incorporated into other workflow management systems for K8s as a docking module. 
We welcome you to download, learn, and work together to maintain CWB with us. If you use it for scientific research and 
engineering applications, please be sure to protect the copyright and indicate authors and source.

##Resource description

###./experiment

This directory includes Argo, Batch Job, and CWB workflow submission deployment files.
Three submission approaches contain resource gathering module, respectively.
You can get the 'usage.txt' of resource gathering module in the directory '/home' on one node when each submission approach is completed.
Our experimental environment consists of one Master and two nodes. In K8s cluster, node name is identified by node's IP.

Master: 192.168.6.109, node1: 192.168.6.110, node2: 192.168.6.111

OS version: CentOS Linux release 7.8/Ubuntu 20.4, Kubernetes: v1.18.6/v1.19.6, Docker: 18.09.6.

If you want to download and run the CWB, you should update the field 'master.ip','name: system:node:' and 'server' included in 
'resourceUsage.yaml', 'rabc-deploy.yaml', 'rabc-argo.yaml', 'buildinjector.yaml', and 'storageClass-nfs.yaml'.

1. ./experiment/argo_test

steps:

##### a. kubectl create ns argo && kubectl apply -f install.yaml -n argo

  The two commands finish the creation of Argo namespace and Argo's installation.
##### b. ./deploy.sh

This shell script file is responsible for deploying RBAC permission and resource gathering module.

##### c. argo submit task-dag.yaml -n argo

Deploy our customized workflow described by Yaml through Argo binary tool.

  argo list -n argo

  argo get -n argo @taskName

  argo logs -n argo @taskName

##### d. ./clean.sh

Clear up RBAC permission and resource gathering module.

##### e. kubectl delete -f install.yaml -n argo

Clear up Argo components.

2. ./experiment/batch_job

steps:

##### ./wfScheduler.sh

The shell script is responsible for deploying the RBAC and resource gathering module and running our 
customized workflow.

3. ./experiment/cwb_test

The directory './deploy' includes the Yaml files corresponding to CWB, RBAC, resource usage rate, Nfs, and workflow injection module.
We use Configmap method in Yaml file to inject workflow information (dependency.json) into the container of workflow injection module.
Refer to './deploy/buildInjector.yaml' for details.

steps:

##### a. ./deploy.sh

Deploy CWB into K8s cluster. This deployment operation includes a series of 'Kubectl' command.

During the workflow lifecycle, you can watch the execution states of workflow tasks. The following is operation command.

'kubectl get pods -A --watch -o wide'

##### b. ./clear.sh

When the workflow is completed, you can run '.clear.sh' file to clean up the workflow information. 

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
