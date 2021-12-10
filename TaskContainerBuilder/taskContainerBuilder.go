package main

import (
	"TaskContainerBuilder/event"
	informer "TaskContainerBuilder/informer"
	k8sResource "TaskContainerBuilder/k8sResource"
	"TaskContainerBuilder/messageProto/TaskContainerBuilder"
	_ "flag"
	_ "fmt"
	_ "github.com/fsnotify/fsnotify"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	_ "k8s.io/apimachinery/pkg/util/runtime"
	_ "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v2 "k8s.io/client-go/listers/core/v1"
	_ "k8s.io/client-go/tools/cache"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)
var WorkflowInjectorServerIp string
var clientset *kubernetes.Clientset
type resourceAllocation struct {
	MilliCpu uint64
	Memory uint64
}
type requestResourceConfig struct {
	Timestamp int64
	AliveStatus bool
	ResourceDemand resourceAllocation
}
type NodeAllocateResource struct {
	MilliCpu uint64
	Memory uint64
}
type NodeUsedResource struct {
	MilliCpu uint64
	Memory uint64
}

type NodeUsedResourceMap map[string]NodeUsedResource
type NodeAllocateResourceMap map[string]NodeAllocateResource
//the structure of workflow task
type WorkflowTask struct {
	//workflow ID
	WorkflowId string
	//taskNum
	TaskNum uint32
	//taskName
	TaskName string
	//task Image
	Image string
	//unit millicore(1Core=1000millicore)
	Cpu uint64
	//unit MiB
	Mem uint64
	//task order in workflow
	TaskOrder uint32
	//env parameters being injected into Pod
	Env map[string]string
	// input Vector
	InputVector []string
	// output Vector
	OutputVector []string
	Args []string
}
var workflowTask WorkflowTask
//type nodeResidualResource struct {
//	MilliCpu int64
//	Memory int64
//}
var nodeResidualResource k8sResource.NodeResidualResource
//type residualResourceMap map[string]nodeResidualResource
var nodeResidualResourceMap k8sResource.ResidualResourceMap
var nodeResidualMap k8sResource.ResidualResourceMap

var allocateResourceMap k8sResource.NodeAllocateResourceMap
var usedResourceMap k8sResource.NodeUsedResourceMap
//var nodeAllocateResourceMap k8sResource.ResidualResourceMap
//var nodeUsedResourceMap k8sResource.ResidualResourceMap

var taskNsNum int64 =0
var taskPodNum int64 = 0

var podLister v2.PodLister
var nodeLister v2.NodeLister
var namespaceLister v2.NamespaceLister
var taskSema = make(chan int, 1)

type  WorkflowTaskMap map[uint32] WorkflowTask
var workflowTaskMap = make(map[uint32] WorkflowTask)
var taskReceive = make(chan int, 1)
var thisTaskPodExist bool

var dependencyMap = make(map[uint32]map[string][]string)
var taskCompletedStateMap = make(map[uint32]bool)

var clusterAllocatedCpu uint64
var clusterAllocatedMemory uint64
var clusterUsedCpu uint64
var clusterUsedMemory uint64
var masterIp string
var gatherTime string
var interval uint32

var experimentalDataObj *os.File

type Builder func(taskObject WorkflowTask) (*TaskContainerBuilder.InputWorkflowTaskResponse,error)

//resource service structure
type ResourceServiceImpl struct {

}

//workflow input interface module, received requests of workflow task generation from
//workflow injection module via gRPC
func (rs *ResourceServiceImpl) InputWorkflowTask(ctx context.Context,request *TaskContainerBuilder.InputWorkflowTaskRequest)(*TaskContainerBuilder.InputWorkflowTaskResponse,error) {
	var response TaskContainerBuilder.InputWorkflowTaskResponse
	log.Printf("--------------The reception time of this workflow task is: %vms",time.Now().UnixNano()/1e6)

	/*Create map[uint32]bool and identify the execution status of each task, initially as false*/
	taskCompletedStateMap[request.TaskOrder] = false

	//Receiving workflow tasks, packed into workflowMap
	taskReceive <- 1
	workflowTaskMap[request.TaskOrder] = WorkflowTask{
		WorkflowId: request.WorkflowId,
		TaskNum: request.TaskNum,
		TaskName:   request.TaskName,
		Image:      request.Image,
		Cpu:        request.Cpu,
		Mem:        request.Mem,
		TaskOrder:  request.TaskOrder,
		Env:        request.Env,
		InputVector: request.InputVector,
		OutputVector: request.OutputVector,
		Args: request.Args,
	}
	<- taskReceive
	log.Println(workflowTaskMap[request.TaskOrder])

	/*If it is the first task in this workflow, invoke CreateTask function to generate task container.
	If not, invoke CreateTask function to generate task container through Update event of Informer.*/
	if request.TaskOrder == 0 && request.InputVector == nil {
		log.Printf("--------------Starting to run the first task [%d] in current workflow is: %vms",request.TaskOrder,time.Now().UnixNano()/1e6)
        //write experimental data into /home/exp.txt.
		//outData := "First task" + strconv.Itoa(int(request.TaskOrder)) + ": " + strconv.Itoa(int(time.Now().UnixNano()/1e6)) + "\n"
		//experimentalDataObj.WriteString(outData)

		res, err := CreateTask(request.TaskOrder)
		if err != nil {
			panic(err.Error())
		}
		log.Printf("This is the first task: %v, order: %v , workflow: %v, InputVector: %v, OuputVector: %v.\n",request.TaskName,
			request.TaskOrder,request.WorkflowId,request.InputVector,request.OutputVector)
		log.Println("*************************************************")
		response = *res
		return &response, nil
	}
	log.Printf("The current taskName: %v: order: %v of workflow: %v.\n",request.TaskName,request.TaskOrder,request.WorkflowId)
	//write experimental data into /home/exp.txt.
	outData := "Receiving " + request.TaskName + ": " + strconv.Itoa(int(time.Now().UnixNano()/1e6)) +"\n"
	experimentalDataObj.WriteString(outData)
	log.Println("--------------------------------------------")
	response.ErrNo = 0
	return &response, nil
}

//Delete workflow namespace
func DeleteWorkflowNamespace(task WorkflowTask) error {
	workflowTaskId := task.WorkflowId
	namespacesClient := clientset.CoreV1().Namespaces()
	deletePolicy := metav1.DeletePropagationForeground
	if err := namespacesClient.Delete(workflowTaskId, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
		return err
	}
	log.Printf("Delete the Namespace: %s.\n", workflowTaskId)
	log.Printf("--------------Delete the Namespace %s in time: %vms.\n", workflowTaskId,time.Now().UnixNano()/1e6)
	//write experimental data into /home/exp.txt.
	outData := "Delete the Namespace " + workflowTaskId + ": " + strconv.Itoa(int(time.Now().UnixNano()/1e6)) + "\n\n"
	experimentalDataObj.WriteString(outData)
	return nil
}

var deletedPodNum int64 = 0
func DeleteCurrentTaskContainer(param interface{})(*TaskContainerBuilder.InputWorkflowTaskResponse,error) {
	var taskPodResponse TaskContainerBuilder.InputWorkflowTaskResponse
	if order, ok := (param).(uint32); ok {
		task := workflowTaskMap[order]
		taskName := task.TaskName

		taskPod, err := podLister.Pods(task.WorkflowId).Get(taskName)
		if err != nil {
			panic(err.Error())
		}
		taskPodHostIp := taskPod.Status.HostIP
		err = clientset.CoreV1().Pods(task.WorkflowId).Delete(taskName, &metav1.DeleteOptions{})
		deletedPodNum++
		log.Printf("Deleting taskPodName: %s.\n", taskName)
		log.Printf("--------------Deleting taskPodName %s in time:%vms on Cluster Node: %v.\n",
			taskName,time.Now().UnixNano()/1e6, taskPodHostIp)

		//write experimental data into /home/exp.txt.
		outData := "Delete " + taskName + ": " + strconv.Itoa(int(time.Now().UnixNano()/1e6)) + " on Cluster Node: "+ taskPodHostIp +"\n"
		experimentalDataObj.WriteString(outData)

		log.Printf("This is the %dth deleted workflow task pod in %v on Cluster node: %v.\n", deletedPodNum,
			task.WorkflowId, taskPodHostIp)
		if err != nil {
			panic(err.Error())
		}
		//log.Println(len(workflowTaskMap))
		if task.TaskOrder == uint32(len(workflowTaskMap)-1) {
			err := DeleteWorkflowNamespace(task)
			if err != nil {
				panic(err.Error())
			}
			//Wake up the next workflow's injection

			workflowTaskMap = make(map[uint32] WorkflowTask)
		}
		/*The previous task has been completed and deleted, and the completed status is set to True.*/
		taskCompletedStateMap[task.TaskOrder] = true
		log.Printf("The bool val of current task[%d] is: %v.\n",task.TaskOrder,
			taskCompletedStateMap[uint32(task.TaskOrder)])
		log.Printf("Staring to trigger the next task container.\n")
		//Trigger a subsequent task
		event.CallEvent("CreateNextTaskContainer",task.TaskOrder)
		taskPodResponse.Result = 1
		return &taskPodResponse, nil
	} else {
		taskPodResponse.Result = 0
		return &taskPodResponse, nil
	}
}
//Delete current Failed task pod
func DeleteCurrentFailedTaskContainer(param interface{})(*TaskContainerBuilder.InputWorkflowTaskResponse,error) {
	var taskPodResponse TaskContainerBuilder.InputWorkflowTaskResponse
	if order, ok := (param).(uint32); ok {
		task := workflowTaskMap[order]
		taskName := task.TaskName

		err := clientset.CoreV1().Pods(task.WorkflowId).Delete(taskName, &metav1.DeleteOptions{})
		deletedPodNum++
		log.Printf("Deleting Failed taskPodName: %s.\n", taskName)
		log.Printf("--------------Deleting Failed taskPodName %s in time:%vms.\n", taskName,time.Now().UnixNano()/1e6)
		log.Printf("This is the %dth deleted workflow task pod in %v.\n", deletedPodNum, task.WorkflowId)
		if err != nil {
			panic(err.Error())
		}

		/*The Failed task has been completed and deleted, and its status is set to False*/
		taskCompletedStateMap[task.TaskOrder] = false
		log.Printf("The bool val of current task[%d] is: %v.\n",task.TaskOrder,
			taskCompletedStateMap[uint32(task.TaskOrder)])
		log.Printf("Staring to trigger the current task container again.\n")
		//The current Failed task is triggered again.
		event.CallEvent("AgainCreateCurrentTaskContainer",task.TaskOrder)
		taskPodResponse.Result = 1
		return &taskPodResponse, nil
	} else {
		taskPodResponse.Result = 0
		return &taskPodResponse, nil
	}
}

//Start to input next workflow task
func WakeUpNextWorkflow(param interface{}) (*TaskContainerBuilder.InputWorkflowTaskResponse,error)  {
	var response TaskContainerBuilder.InputWorkflowTaskResponse
	/*Obtan the serviceName of workflow injection module*/
	WorkflowInjectorServer := os.Getenv("WORKFLOW_INJECTOR_SERVICE_HOST")
	WorkflowInjectorPort := os.Getenv("WORKFLOW_INJECTOR_SERVICE_PORT")
	WorkflowInjectorServerIp = WorkflowInjectorServer + ":" + WorkflowInjectorPort
	//WorkflowInjectorServerIp = "192.168.6.111:7070"
	log.Println(WorkflowInjectorServerIp)
	if wfIndex, ok := (param).(uint32); ok {
		//Dial and connect
		log.Println("Dial Workflow Injector...")
		//WorkflowInjectorServerIp :="192.168.6.111:7070"
		conn, err := grpc.Dial(WorkflowInjectorServerIp, grpc.WithInsecure())
		if err != nil {
			panic(err.Error())
		}
		defer conn.Close()
		//Create client instance
		visitWorkflowInjectorClient := TaskContainerBuilder.NewWorkflowInjectorServiceClient(conn)
		//FinishedWorkflowId:string(strconv.Atoi(task.WorkflowId)+1)
		responseInfo := &TaskContainerBuilder.NextWorkflowSendRequest{
			FinishedWorkflowId: strconv.Itoa(int(wfIndex)),
		}
		requestNextWorkflowResponse, err := visitWorkflowInjectorClient.NextWorkflowSend(context.Background(), responseInfo)
		if err != nil {
			panic(err.Error())
		}
		if requestNextWorkflowResponse.Result == true{
			log.Println(requestNextWorkflowResponse.Result)
			log.Println("--------------The next workflow is injected at time:",time.Now().UnixNano()/1e6)
			//write experimental data into /home/exp.txt.
			outData := "Injecting workflow" + strconv.Itoa(int(wfIndex)+1)+"is over"+ ": " + strconv.Itoa(int(time.Now().UnixNano()/1e6)) +"\n"
			experimentalDataObj.WriteString(outData)

		} else {
			log.Println("Task container builder's work is over.")
		}
		response.Result = 1
		return &response, nil
	} else {
		response.Result = 0
		return &response, nil
	}
}

/*Invoke ScheduleNextBatchTaskContainer()
Check the OutputVector of the currently completed task. If it is empty, it will be the last task in the workflow
and trigger the next workflow.
If not empty, the function fetches the task sequence number and calls the pod builder concurrently.*/
func ScheduleNextBatchTaskContainer(param interface{})(*TaskContainerBuilder.InputWorkflowTaskResponse,error) {
	var response TaskContainerBuilder.InputWorkflowTaskResponse
	wait := sync.WaitGroup{}
	log.Println("Enter ScheduleNextBatchTaskContainer function...")
	if order, ok := (param).(uint32); ok {
		if task, ok := workflowTaskMap[order]; ok {
			for _, output := range task.OutputVector {
				index, err := strconv.Atoi(output)
				if err != nil {
					panic(err.Error())
				}
				log.Printf("The outVector of current task is: [%d]\n",index)
				log.Println("--------------Start time for goroutine in CreateTaskforScheduleNextBatch in time:", time.Now().UnixNano()/1e6)
				wait.Add(1)
				go CreateTaskforScheduleNextBatch(uint32(index),&wait)
				//go CreateTaskforScheduleNextBatch(uint32(index))
			}
//			defer runtime.HandleCrash()
			wait.Wait()
			response.Result = 1
			return &response, nil
		} else{
			response.Result = 1
			response.VolumePath = ""
			response.ErrNo = 0
			return &response, nil
		}
	}else{
		response.Result = 1
		response.VolumePath = ""
		response.ErrNo = 0
		return &response, nil
	}
}
func CreateTaskforScheduleNextBatch(order uint32,waiter *sync.WaitGroup) {
//func CreateTaskforScheduleNextBatch(order uint32) {

	defer waiter.Done()
	log.Println("Enter CreateTaskforScheduleNextBatch function...")
		if task, ok := workflowTaskMap[order]; ok {
			/*When this task is the first task of current workflow.*/
			if task.InputVector == nil {
				taskSema <- 1
				_, _, err := CreateTaskContainer(task,clientset)
				<-taskSema
				if err != nil {
					panic(err.Error())
					//return  volumePath, errNo, err
				}
				//return volumePath, errNo, nil
			} else{
				/* Check InputVector*/
				accumulatedBoolVar := true
				for _, input := range task.InputVector {
					index, err := strconv.Atoi(input)
					if err != nil {
						panic(err.Error())
					}
					log.Printf("The intVector of current task is: [%d]\n",index)
					log.Printf("The bool val of task[%d] is: %v.\n",index,taskCompletedStateMap[uint32(index)])
					accumulatedBoolVar = accumulatedBoolVar && taskCompletedStateMap[uint32(index)]
				}
				log.Printf("The intVector accumulatedBoolVar of current task is: [%v]\n",accumulatedBoolVar)
				/*Create a pod for this task when all the previous tasks of the current task have completed and the
				current task have not been executed*/
				if accumulatedBoolVar && (!taskCompletedStateMap[task.TaskOrder]) {
					log.Println("Prefix task of current task are all completed.")
					log.Printf("--------------Starting to run task[%d] in time:%vms.\n",task.TaskOrder,time.Now().UnixNano()/1e6)
					taskSema <- 1
					_, _, err := CreateTaskContainer(task,clientset)
					<-taskSema
					if err != nil {
						panic(err.Error())
						//return volumePath, errNo, err
					}
					//return volumePath, errNo, err
				}
				//return "", 0 ,nil
			}
		}
}
/*Create task container*/
func CreateTask(param interface{})(*TaskContainerBuilder.InputWorkflowTaskResponse,error) {
	var taskPodResponse TaskContainerBuilder.InputWorkflowTaskResponse

	if order, ok := (param).(uint32); ok {
		if task, ok := workflowTaskMap[order]; ok {
			if task.InputVector == nil {
				taskSema <- 1
				volumePath, errNo, err := CreateTaskContainer(task,clientset)
				<-taskSema

				taskPodResponse.VolumePath = volumePath
				if err != nil {
					panic(err.Error())
					taskPodResponse.ErrNo = errNo
					return &taskPodResponse, nil
				}
				taskPodResponse.Result = 0
				return &taskPodResponse, nil
			} else{
				accumulatedBoolVar := true
				for _, input := range task.InputVector {
					index, err := strconv.Atoi(input)
					if err != nil {
						panic(err.Error())
					}
					accumulatedBoolVar = accumulatedBoolVar && taskCompletedStateMap[uint32(index)]
				}
				if accumulatedBoolVar && (!taskCompletedStateMap[task.TaskOrder]) {
					log.Println("Prefix task of current task are all completed.")
					taskSema <- 1
					volumePath, errNo, err := CreateTaskContainer(task,clientset)
					<-taskSema

					taskPodResponse.VolumePath = volumePath
					if err != nil {
						panic(err.Error())
						taskPodResponse.ErrNo = errNo
						return &taskPodResponse, nil
					}
					taskPodResponse.Result = 0
					return &taskPodResponse, nil

				} else {
					taskPodResponse.Result = 1
					return &taskPodResponse, nil
				}
			}

		} else{
			taskPodResponse.Result = 1
			return &taskPodResponse, nil
		}

	} else {
		taskPodResponse.Result = 1
		return &taskPodResponse, nil
	}
}

func CreateTaskContainer(task WorkflowTask,clientService *kubernetes.Clientset)(string, uint32, error) {
	//taskSema <- 1
	//clientNamespace, isFirstPod, clientPvcOfThisNamespace, err := createTaskPodNamespaces(request, clientset)

	clientNamespace, isFirstPod, clientPvcOfThisNamespace, err := createTaskPodNamespaces(task, clientService)
	if err != nil {
		panic(err.Error())
	}
	//<-taskSema

	//volumePath, err := clientTaskCreatePod(request, clientset, clientNamespace, isFirstPod, clientPvcOfThisNamespace)
	volumePath, errNo, err := clientTaskCreatePod(task, clientService, clientNamespace, isFirstPod, clientPvcOfThisNamespace)

	return  volumePath,errNo, err
}


func recoverNamespaceListerFail() {
	if r := recover(); r!= nil {
		log.Println("recovered from createTaskPodNamespaces()", r)
	}
}

//Create namespace
func createTaskPodNamespaces(task WorkflowTask,clientService *kubernetes.Clientset)(*v1.Namespace, bool,
	*v1.PersistentVolumeClaim, error) {
	defer recoverNamespaceListerFail()
	name := task.WorkflowId
	namespacesClient := clientService.CoreV1().Namespaces()
	/*3.1 Monitor Namespace resources and obtain namespaceLister using the Informer tool package*/
	namespaceList, err := namespaceLister.List(labels.Everything())
	if err != nil {
		log.Println(err)
		panic(err.Error())
	}
	namespaceExist := false
	for _, ns := range namespaceList{
		if ns.Name == name {
			namespaceExist = true
			break
		}
	}
	// Create pvc's client
	pvcClient := clientService.CoreV1().PersistentVolumeClaims(name)
	/*When the Namespace of this workflow exists...*/
	if namespaceExist {
		nsClientObject ,err := namespacesClient.Get(name,metav1.GetOptions{})
		if err != nil {
			panic(err)
		}
		log.Printf("This namespace %v is already exist.\n", name)
		pvcObjectAlreadyInThisNamespace, err := pvcClient.Get(name+"-pvc", metav1.GetOptions{})
		//The namespace is already exist and there have been a pod in this namespace, podIsFirstInThisNamespace
		//condition is false
		podIsFirstInThisNamespace := false
		return nsClientObject, podIsFirstInThisNamespace, pvcObjectAlreadyInThisNamespace,nil
	}else{
		taskNsNum++
		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{"namespace": "task"},
			},
			Status: v1.NamespaceStatus{
				Phase: v1.NamespaceActive,
			},
		}
		//write experimental data into /home/exp.txt.
		outData := "Create" + name + ": " + strconv.Itoa(int(time.Now().UnixNano()/1e6)) + "\n"
		experimentalDataObj.WriteString(outData)
		// Create a new Namespaces
		clientNamespaceObject, err := namespacesClient.Create(namespace)
		podIsFirstInThisNamespace := true
		if err != nil {
			panic(err)
		}
		log.Printf("Creating Namespace %s\n", clientNamespaceObject.ObjectMeta.Name)
		log.Printf("Creating Namespaces...,this is %dth namespace.\n",taskNsNum)
		//Create namespace's pvc

		pvcObjectInThisNamespace,err := createPvc(clientService,name)
		if err != nil {
			panic(err)
		}
		return clientNamespaceObject,podIsFirstInThisNamespace, pvcObjectInThisNamespace, nil
		//return clientNamespaceObject,podIsFirstInThisNamespace, nil
	}
}
func recoverCreatePvcFail() {
	if r := recover(); r!= nil {
		log.Println("recovered from ", r)
	}
}
func createPvc(clientService *kubernetes.Clientset,nameOfNamespace string)(*v1.PersistentVolumeClaim,error) {
	defer recoverCreatePvcFail()
	storageClassName := "nfs-storage"
	pvcClient := clientService.CoreV1().PersistentVolumeClaims(nameOfNamespace)
	pvc := new(v1.PersistentVolumeClaim)
	pvc.TypeMeta = metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"}
	pvc.ObjectMeta = metav1.ObjectMeta{ Name: nameOfNamespace +"-pvc",
	//Create namespace's pvc
		Labels: map[string]string{"pvc": "nfs"},
		Namespace: nameOfNamespace }

		pvc.Spec = v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			Resources: v1.ResourceRequirements{
				Requests:v1.ResourceList{
					v1.ResourceStorage: resource.MustParse("10" + "Mi"),
				},
			},
			//Selector: &metav1.LabelSelector{
			//	MatchLabels:
			//	map[string]string{"app": "nfs"},
			//},
			StorageClassName: &storageClassName,
		}
	//}
	//write experimental data into /home/exp.txt.
	outData := "Create " + pvc.Name + ": " + strconv.Itoa(int(time.Now().UnixNano()/1e6)) + "\n"
	experimentalDataObj.WriteString(outData)

	pvcResult, err := pvcClient.Create(pvc)
	if err != nil {
		panic(err)
	}
	log.Printf("Created Pvc %s on %s\n", pvcResult.ObjectMeta.Name,
		pvcResult.ObjectMeta.CreationTimestamp)

	return pvcResult, nil
}

func recoverCreatePodFail() {
	if r := recover(); r!= nil {
		log.Println("recovered from ", r)
	}
}
//4. Task container creator module
func clientTaskCreatePod(task WorkflowTask ,clientService *kubernetes.Clientset,
	podNamespace *v1.Namespace, IsfirstPod bool, pvcClient *v1.PersistentVolumeClaim)(string, uint32, error) {
	defer recoverCreatePodFail()
	//schedulerName := "default-scheduler"
	taskPodNum++
	//taskPodName := task.WorkflowId + "-task-"+strconv.Itoa(int(task.TaskOrder))
	taskPodName := task.TaskName
	taskPod, err := podLister.Pods(task.WorkflowId).Get(taskPodName)
	if err == nil {
		thisTaskPodExist = true
		log.Printf("This task pod: %v is already exist in %v.\n", taskPod.Name, taskPod.Namespace)
	} else {
		//panic(err)
		/*err不为nil,说明Informer工具包的podLister不能获取到，此任务pod不存在*/
		thisTaskPodExist = false
	}

	if thisTaskPodExist == true {
		volumePathInContainer := "/nfs/data"
		event.CallEvent("DeleteCurrentTaskContainer", task.TaskOrder)
		returnErrNo := uint32(1)
		return volumePathInContainer, returnErrNo, nil

	} else {
		//When the remaining resource suffices for request resource of task Pod, break the for loop.
		for {
			//Obtain the map of remaining resources of each node
			ResidualMap := k8sResource.GetK8sEachNodeResource(podLister, nodeLister, nodeResidualMap)

			cpuResidualResourceMinValue := uint64(0)
			memResidualResourceMinValue := uint64(0)
			//Obtain the largest remaining resources in the cluster nodes
			for _, val := range ResidualMap {
				if (val.MilliCpu > cpuResidualResourceMinValue) && (val.MilliCpu != 0) {
					cpuResidualResourceMinValue = val.MilliCpu
				}
				if (val.Memory > memResidualResourceMinValue) && (val.Memory != 0) {
					memResidualResourceMinValue = val.Memory
				}
			}
			/*If the maximum remaining resources of a cluster node are greater than the resources required by the current
			task, a task container would be generated*/
			if (cpuResidualResourceMinValue >= task.Cpu) && (memResidualResourceMinValue >= task.Mem) {
				//Create this task POD if the cluster resources are sufficient
				break
			}
			log.Println("Remaining resource does not suffices for creating task pods and execute for loop.")
			//else {
			//	return "", 0, nil
			//}
		}
		log.Println("Remaining resource suffices for creating task pods.")
		path, returnErrNo, err := CreateTaskPod(task, clientService, podNamespace, IsfirstPod, pvcClient)
		VolumePath := path
		if err != nil {
			panic(err.Error())
		}
		return VolumePath, returnErrNo, nil
	}
}

func CreateTaskPod(task WorkflowTask ,clientService *kubernetes.Clientset,
	podNamespace *v1.Namespace, IsfirstPod bool, pvcClient *v1.PersistentVolumeClaim)(string, uint32, error){
	if IsfirstPod {
		//hostPath := "/nfs/data/"
	}
	pod := new(v1.Pod)
	//pod.TypeMeta = unversioned.TypeMeta{Kind: "Pod", APIVersion: "v1"}
	pod.TypeMeta =  metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"}
	pod.ObjectMeta = metav1.ObjectMeta{
		//Name: task.WorkflowId + "-" + strconv.Itoa(int(task.TaskOrder)),
		Name: task.TaskName,
		Namespace: podNamespace.Name,
		Labels: map[string]string{"app": "task"},
		Annotations: map[string]string{"AnnotationsName": "task pods of workflow."}}
	volumePathInContainer := "/nfs/data"

	pod.Spec = v1.PodSpec{
		RestartPolicy: v1.RestartPolicyNever,
		SchedulerName: "default-scheduler",
		Volumes: []v1.Volume{
			v1.Volume{
				Name: "pod-share-volume",
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvcClient.ObjectMeta.Name,
					},
				},
			},
		},
		Containers: []v1.Container{
			v1.Container{
				Name:    "task-pod",
				Image:   task.Image,
				Command: nil,
				//Command: nil,
				//Args:[]string{"-c","1","-m","100","-t","5","-i","3"},
				Args: task.Args,
				Ports: []v1.ContainerPort{
					v1.ContainerPort{
						ContainerPort: 80,
						Protocol:      v1.ProtocolTCP,
					},
				},
				VolumeMounts: []v1.VolumeMount {
					v1.VolumeMount{
						Name: "pod-share-volume",
						MountPath: volumePathInContainer,
						//SubPath: pvcClient.ObjectMeta.Name,
					},
				},
				Env: []v1.EnvVar{
					v1.EnvVar{
						Name:  "VOLUME_PATH",
						Value: volumePathInContainer,
						//ValueFrom: &v1.EnvVarSource{
						//	FieldRef: &v1.ObjectFieldSelector{
						//		FieldPath: volumePathInContainer,
						//	},
					},
				},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse(strconv.Itoa(int(task.Cpu)) + "m"),
						v1.ResourceMemory: resource.MustParse(strconv.Itoa(int(task.Mem)) + "Mi"),
					},
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse(strconv.Itoa(int(task.Cpu)) + "m"),
						v1.ResourceMemory: resource.MustParse(strconv.Itoa(int(task.Mem)) + "Mi"),
					},
				},
				ImagePullPolicy: v1.PullIfNotPresent,
				//ImagePullPolicy: v1.PullAlways,
			},
			},
		}
	//write experimental data into /home/exp.txt.
	outData := "Create " + pod.Name + ": " + strconv.Itoa(int(time.Now().UnixNano()/1e6)) + "\n"
	experimentalDataObj.WriteString(outData)

	_, err := clientService.CoreV1().Pods(podNamespace.Name).Create(pod)
	if err != nil {
		panic(err.Error())
	}
	//namespace := podNamespace.Name
	pods, err := clientService.CoreV1().Pods(podNamespace.Name).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
		//return "", err
	}
	log.Printf("There are %d pods in current namespace:%s\n", len(pods.Items), task.WorkflowId)
	log.Printf("This is %dth task pod in all workflow namespaces.\n", taskPodNum)
	//for _, pod := range pods.Items {
	//	log.Printf("Name: %s, Status: %s, CreateTime: %s\n", pod.ObjectMeta.Name, pod.Status.Phase, pod.ObjectMeta.CreationTimestamp)
	//}
	errNo := uint32(0)
	return volumePathInContainer,errNo, nil
}

func taskBuilderServer(waiter *sync.WaitGroup) {
	defer waiter.Done()
	//Build grpc server
	server := grpc.NewServer()
	log.Println("Build workflow task builder gRPC Server.")
	//Register the resource request service
	TaskContainerBuilder.RegisterTaskContainerBuilderServiceServer(server,new(ResourceServiceImpl))
	//Listen on 7070 port
	lis, err := net.Listen("tcp", ":7070")
	log.Println("Listening local port 7070.")
	if err != nil {
		panic(err.Error())
	}
	server.Serve(lis)
}

//Initialize the nodeResidualResourceMap
func initNodeResidualResourceMap(resourceMap k8sResource.ResidualResourceMap, clusterMasterIp string) k8sResource.ResidualResourceMap {
	//nodeIpPrefix := "121.250.173."
	splitName :=  strings.Split(clusterMasterIp,".")
	nodeIpFourthField, err := strconv.Atoi(splitName[len(splitName)-1])
	if err != nil {
		panic(err)
	}
	nodeNum, err := strconv.Atoi(os.Getenv("NODE_NUM"))
	if err != nil {
		panic(err)
	}
	nodeIpThirdField, err := strconv.Atoi(splitName[2])
	if err != nil {
		panic(err)
	}
	nodeIpPrefix := splitName[0] + "." +splitName[1] + "." + splitName[2] + "."

	for i := 1; i <= nodeNum; i++ {
		if (nodeIpFourthField+i) < 256 {
			nodeResidualResourceKey := nodeIpPrefix + strconv.Itoa( nodeIpFourthField+i )
			resourceMap[nodeResidualResourceKey] = k8sResource.NodeResidualResource{0, 0}
		}else {
			nodeIpFourthField = nodeIpFourthField + i - 256
			nodeIpThirdField = nodeIpThirdField + 1
			nodeIpPrefix = splitName[0] + "." +splitName[1] + "." + strconv.Itoa(nodeIpThirdField) + "."
			nodeResidualResourceKey := nodeIpPrefix + strconv.Itoa(nodeIpFourthField+i)
			resourceMap[nodeResidualResourceKey] = k8sResource.NodeResidualResource{0, 0}
		}
	}
	log.Println(resourceMap)
	return resourceMap
}
//Event trigger module
func eventRegister(){
	// Register event
	event.RegisterEvent("CreateNextTaskContainer", ScheduleNextBatchTaskContainer)
	event.RegisterEvent("AgainCreateCurrentTaskContainer",CreateTask)
	event.RegisterEvent("ThisWorkflowEnd", WakeUpNextWorkflow)
	event.RegisterEvent("DeleteCurrentTaskContainer", DeleteCurrentTaskContainer)
	event.RegisterEvent("DeleteCurrentFailedTaskContainer",DeleteCurrentFailedTaskContainer)
}
func main() {
	taskCompletedStateMap = make(map[uint32]bool)
	logFile, err := os.OpenFile("/home/log.txt", os.O_CREATE | os.O_APPEND | os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout,logFile)
	log.SetOutput(mw)
	//create experimental data file
	experimentalDataObj, _ = os.OpenFile("/home/exp.txt", os.O_CREATE | os.O_APPEND | os.O_RDWR, 0666)
	defer experimentalDataObj.Close()

	//Get MasterIp by env
	masterIp = os.Getenv("MASTER_IP")
	log.Printf("masterIp: %v\n",masterIp)
	//Get time interval of sample by env
	gatherTime = os.Getenv("GATHER_TIME")
	log.Printf("gatherTime: %v\n",gatherTime)
	valTime, err := strconv.Atoi(gatherTime)
	if err != nil {
		panic(err)
	}
	interval = uint32(valTime)

	//Create chan for Informer
	stopper := make(chan struct{})
	defer close(stopper)
    /*Create a map of remaining resources for each node*/
	nodeResidualResourceMap = make(k8sResource.ResidualResourceMap)
	nodeResidualMap = initNodeResidualResourceMap(nodeResidualResourceMap, masterIp)

	waiter := sync.WaitGroup{}
	waiter.Add(2)

	//Start gRPC Server
	go taskBuilderServer(&waiter)
	//Start event register
	eventRegister()
	//Create K8s's client
	clientset = k8sResource.GetRemoteK8sClient()
	podLister, nodeLister, namespaceLister = informer.InitInformer(stopper,"/kubelet.kubeconfig")

	//Collect allocatable and request resources periodically
	//go gatherResource(&waiter,allocateResourceMap,usedResourceMap,interval)

	defer runtime.HandleCrash()

	waiter.Wait()
}
