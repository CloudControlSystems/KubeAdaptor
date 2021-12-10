package informer

import (

	"TaskContainerBuilder/event"
	"TaskContainerBuilder/k8sResource"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"log"
	"strconv"
	"strings"
	"time"
)
//the flag of completed workflow
type CurrentWorkflowFinished bool
type CurrentWorkflowTaskFinished bool
var  currentWorkflowFinished CurrentWorkflowFinished
var  currentWorkflowTaskFinished CurrentWorkflowTaskFinished
//var task taskContainerBuilder.WorkflowTask

//State tracking and resource monitoring module
func InitInformer(stop chan struct{}, configfile string) (v1.PodLister,v1.NodeLister,v1.NamespaceLister){


	currentWorkflowFinished = false
	currentWorkflowTaskFinished = false

	//Connect K8s's apiserver
	informerClientset := k8sResource.GetInformerK8sClient(configfile)
	//Innitialize informer
	factory := informers.NewSharedInformerFactory(informerClientset, time.Second*1)

	//Create podInformer
	podInformer := factory.Core().V1().Pods()
	informerPod := podInformer.Informer()

	//Create nodeInformer
	nodeInformer := factory.Core().V1().Nodes()
	informerNode := nodeInformer.Informer()

	//Create namespaceInformer
	namespaceInformer := factory.Core().V1().Namespaces()
	informerNamespace := namespaceInformer.Informer()

	//Create Listers
	podInformerLister := podInformer.Lister()
	nodeInformerLister := nodeInformer.Lister()
	namespaceInformerLister := namespaceInformer.Lister()

	go factory.Start(stop)

	//Synchronize podList
	if !cache.WaitForCacheSync(stop, informerPod.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return nil, nil, nil
	}
	//Synchronize nodeList
	if !cache.WaitForCacheSync(stop, informerNode.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return nil, nil, nil
	}
	//Synchronize namespaceList
	if !cache.WaitForCacheSync(stop, informerNamespace.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return nil, nil, nil
	}

	informerPod.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onPodAdd,
		UpdateFunc: onPodUpdate,
		DeleteFunc: onPodDelete,
	})

	informerNode.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onNodeAdd,
		UpdateFunc: onNodeUpdate,
		DeleteFunc: onNodeDelete,
	})

	informerNamespace.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onNamespaceAdd,
		UpdateFunc: onNamespaceUpdate,
		DeleteFunc: onNamespaceDelete,
	})

	return podInformerLister, nodeInformerLister, namespaceInformerLister
}
func onPodAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)
	//log.Println("add a pod:", pod.Name)
	log.Printf("--------------add a pod[%s] in time:%v.\n", pod.Name,time.Now().UnixNano()/1e6)
}

func onPodUpdate(old interface{}, current interface{}) {

	oldpod := old.(*corev1.Pod)
	//fmt.Println(pod.Name)
	//fmt.Println(pod.Status.Phase)
	oldstatus := oldpod.Status.Phase
	splitName :=  strings.Split(oldpod.Name,"-")
	newpod := current.(*corev1.Pod)
	//fmt.Println(newpod.Name)
	//fmt.Println(newpod.Status.Phase)
	newstatus := newpod.Status.Phase
	if ((oldstatus == "Pending") || (oldstatus == "Running")) && (newstatus == "Succeeded") {
		log.Println(oldpod.Status.Phase)
		log.Println(newpod.Status.Phase)

		log.Printf("%v: pod is completed", oldpod.Name)
		taskOrder, err := strconv.Atoi(splitName[3])
		if err != nil {
			panic(err.Error())
		}
		wfTaskIndex := uint32(taskOrder)
		//Delete successful task pod and start the next task pod.
		event.CallEvent("DeleteCurrentTaskContainer", wfTaskIndex)

		//if oldpod.Name[0:8] == "workflow" {
		//	log.Printf("%v: pod is completed", oldpod.Name)
		//	taskOrder, err := strconv.Atoi(oldpod.Name[16:])
		//	//log.Println(taskOrder)
		//	if err != nil {
		//		panic(err.Error())
		//	}
		//	wfTaskIndex := uint32(taskOrder)
		//	//Delete successful task pod and start the next task pod.
		//	event.CallEvent("DeleteCurrentTaskContainer", wfTaskIndex)
		//}
		log.Println("----------------------------")
	} else if newstatus == "Failed" {
		log.Println(oldpod.Status.Phase)
		log.Println(newpod.Status.Phase)
		log.Println("Pod error.", oldpod.Name)
		taskOrder, err := strconv.Atoi(splitName[3])
		if err != nil {
			panic(err.Error())
		}
		wfTaskStruct := uint32(taskOrder)
		event.CallEvent("DeleteCurrentFailedTaskContainer", wfTaskStruct)

	}
}
func onPodDelete(obj interface{}) {
	pod := obj.(*corev1.Pod)
	//log.Println("delete a pod:", pod.Name)
	log.Println("--------------delete a pod at time:", pod.Name,time.Now().UnixNano()/1e6)
}
func onNodeAdd(obj interface{}) {
	node := obj.(*corev1.Node)
	log.Println("add a node:", node.Name)
}
func onNodeUpdate(old interface{}, current interface{}) {
	//log.Println("updating..............")
}
func onNodeDelete(obj interface{}) {
	node := obj.(*corev1.Node)
	log.Println("delete a Node:", node.Name)
}
func onNamespaceAdd(obj interface{}) {
	namespace := obj.(*corev1.Namespace)
	log.Println("add a namespace:", namespace.Name)
}
func onNamespaceUpdate(old interface{}, current interface{}) {
	//log.Println("updating..............")
	//oldNamespace := old.(*corev1.Namespace)
	//oldstatus := oldNamespace.Status.Phase
	//log.Println(oldNamespace.Status.Phase)
}
func onNamespaceDelete(obj interface{}) {
	namespace := obj.(*corev1.Namespace)
	//log.Println("Delete a namespace:", namespace.Name)
	log.Println("--------------delete a namespace at time:", namespace.Name,time.Now().UnixNano()/1e6)
	splitName :=  strings.Split(namespace.Name,"-")
	wfIndex, err := strconv.Atoi(splitName[1])
	//log.Println(wfIndex)
	if err != nil {
		panic(err.Error())
	}
	//inform that the last task is done
	event.CallEvent("ThisWorkflowEnd",uint32(wfIndex))
}
