package main

import (
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type resourceRequest struct {
	MilliCpu uint64
	Memory uint64
	EphemeralStorage uint64
}
type NodeResidualResource struct {
	MilliCpu uint64
	Memory uint64
}
type NodeAllocateResource struct {
	MilliCpu uint64
	Memory uint64
}
type NodeUsedResource struct {
	MilliCpu uint64
	Memory uint64
}

type ResidualResourceMap map[string]NodeResidualResource
type NodeUsedResourceMap map[string]NodeUsedResource
type NodeAllocateResourceMap map[string]NodeAllocateResource

var resourceRequestNum resourceRequest
var resourceAllocatableNum resourceRequest

var podLister v1.PodLister
var nodeLister v1.NodeLister
var namespaceLister v1.NamespaceLister

var clientset *kubernetes.Clientset

var clusterAllocatedCpu uint64
var clusterAllocatedMemory uint64
var clusterUsedCpu uint64
var clusterUsedMemory uint64
var masterIp string
var gatherTime string
var interval uint32

func GetRemoteK8sClient() *kubernetes.Clientset {
	//k8sconfig= flag.String("k8sconfig","/opt/kubernetes/cfg/kubelet.kubeconfig","kubernetes config file path")
	//flag.Parse()
	//var k8sconfig string
	k8sconfig, err  := filepath.Abs(filepath.Dir("/etc/kubernetes/kubelet.kubeconfig"))
	if err != nil {
		panic(err.Error())
	}
	config, err := clientcmd.BuildConfigFromFlags("",k8sconfig+ "/kubelet.kubeconfig")
	if err != nil {
		log.Println(err)
	}
	//viper.AddConfigPath("/opt/kubernetes/cfg/")     //设置读取的文件路径
	//viper.SetConfigName("kubelet") //设置读取的文件名
	//viper.SetConfigType("yaml")   //设置文件的类型
	//k8sconfig := viper.ReadInConfig()
	//viper.WatchConfig()
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Println("Connect k8s success.")
	}
	return clientset
}

func GetInformerK8sClient(configfile string) *kubernetes.Clientset {
	//k8sconfig= flag.String("k8sconfig","/opt/kubernetes/cfg/kubelet.kubeconfig","kubernetes config file path")
	//flag.Parse()
	//var k8sconfig string
	//if configfile == "/kubelet.kubeconfig" {
	//k8sconfig, err  := filepath.Abs(filepath.Dir("/opt/kubernetes/cfg/kubelet.kubeconfig"))

	k8sconfig, err  := filepath.Abs(filepath.Dir("/etc/kubernetes/kubelet.kubeconfig"))
	if err != nil {
		panic(err.Error())
	}
	config, err := clientcmd.BuildConfigFromFlags("",k8sconfig + configfile)
	if err != nil {
		log.Println(err)
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Println(configfile)
		log.Println("Connect this cluster's k8s successfully.")
	}
	return clientset
}


func InitInformer(stop chan struct{}, configfile string) (v1.PodLister,v1.NodeLister,v1.NamespaceLister){

	//Connect apiserver of K8s and create clientset
	informerClientset := GetInformerK8sClient(configfile)
	//Initialize informer
	factory := informers.NewSharedInformerFactory(informerClientset, time.Second*1)

	//Create podInformer
	podInformer := factory.Core().V1().Pods()
	informerPod := podInformer.Informer()

	// Create nodeInformer
	nodeInformer := factory.Core().V1().Nodes()
	informerNode := nodeInformer.Informer()

	//Create namespaceInformer
	namespaceInformer := factory.Core().V1().Namespaces()
	informerNamespace := namespaceInformer.Informer()

	//Create podLister, nodeLister and namespaceLister
	podInformerLister := podInformer.Lister()
	nodeInformerLister := nodeInformer.Lister()
	namespaceInformerLister := namespaceInformer.Lister()

	//Run all resource objects cache.SharedIndexInformer
	go factory.Start(stop)

	//Syncronize pod resources with apiserver
	if !cache.WaitForCacheSync(stop, informerPod.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return nil, nil, nil
	}
	//Syncronize node resources with apiserver
	if !cache.WaitForCacheSync(stop, informerNode.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return nil, nil, nil
	}
	//Syncronize namespace resources with apiserver
	if !cache.WaitForCacheSync(stop, informerNamespace.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return nil, nil, nil
	}
	// Use customized pod events handler
	informerPod.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onPodAdd,
		UpdateFunc: onPodUpdate,
		DeleteFunc: onPodDelete,
	})
	// Use customized node events handler
	informerNode.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onNodeAdd,
		UpdateFunc: onNodeUpdate,
		DeleteFunc: onNodeDelete,
	})
	// Use customized namespace events handler
	informerNamespace.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onNamespaceAdd,
		UpdateFunc: onNamespaceUpdate,
		DeleteFunc: onNamespaceDelete,
	})

	return podInformerLister, nodeInformerLister, namespaceInformerLister
}
func onPodAdd(obj interface{}) {
	//pod := obj.(*corev1.Pod)
	////log.Println("add a pod:", pod.Name)
	//log.Printf("--------------add a pod[%s] in time:%v.\n", pod.Name,time.Now().UnixNano()/1e6)
}

func onPodUpdate(old interface{}, current interface{}) {
}
func onPodDelete(obj interface{}) {
	//pod := obj.(*corev1.Pod)
	////log.Println("delete a pod:", pod.Name)
	//log.Println("--------------delete a pod at time:", pod.Name,time.Now().UnixNano()/1e6)
}
func onNodeAdd(obj interface{}) {
	//node := obj.(*corev1.Node)
	//log.Println("add a node:", node.Name)
}
func onNodeUpdate(old interface{}, current interface{}) {
	//log.Println("updating..............")
}
func onNodeDelete(obj interface{}) {
	//node := obj.(*corev1.Node)
	//log.Println("delete a Node:", node.Name)
}
func onNamespaceAdd(obj interface{}) {
	//namespace := obj.(*corev1.Namespace)
	//log.Println("add a namespace:", namespace.Name)
}
func onNamespaceUpdate(old interface{}, current interface{}) {
	//log.Println("updating..............")
	//oldNamespace := old.(*corev1.Namespace)
	//oldstatus := oldNamespace.Status.Phase
	//log.Println(oldNamespace.Status.Phase)
}
func onNamespaceDelete(obj interface{}) {
}

//Traverse all hte pods of Nodes in K8s Cluster and obtain the account of request of all pods
func getEachNodePodsResourceRequest(pods []*corev1.Pod,nodeName string) resourceRequest {
	resourceRequestNum = resourceRequest{0, 0, 0}
	for _, pod := range pods {
		if pod.Status.HostIP == nodeName {
			if (pod.Status.Phase == "Running")||(pod.Status.Phase == "Pending"){
				for _, container := range pod.Spec.Containers {
					resourceRequestNum.MilliCpu += uint64(container.Resources.Requests.Cpu().MilliValue())
					resourceRequestNum.Memory += uint64(container.Resources.Requests.Memory().Value())
					resourceRequestNum.EphemeralStorage += uint64(container.Resources.Requests.StorageEphemeral().Value())
				}
				for _, initContainer := range pod.Spec.InitContainers {
					resourceRequestNum.MilliCpu += uint64(initContainer.Resources.Requests.Cpu().MilliValue())
					resourceRequestNum.Memory += uint64(initContainer.Resources.Requests.Memory().Value())
				}
				//fmt.Printf("cpu = %d\n",resourceRequestNum.MilliCpu)
				//fmt.Printf("mem = %d\n",resourceRequestNum.Memory/1024/1024)
			}
		}
		//log.Printf("This %s's HostIP is %s .\n",pod.Name,pod.Status.HostIP)
	}
	return resourceRequestNum
}
// obtain the account of allocatable of all nodes in cluster
func getEachNodeAllocatableNum( nodes []*corev1.Node, nodeName string) resourceRequest {
	resourceAllocatableNum = resourceRequest{0,0,0}
	for _, nod := range nodes {
		if nod.Name == nodeName {
			//if nod.Name == "k8s4-node1" && nodeName == "192.168.6.110" {
			resourceAllocatableNum.MilliCpu = uint64(nod.Status.Allocatable.Cpu().MilliValue())
			resourceAllocatableNum.Memory = uint64(nod.Status.Allocatable.Memory().Value())
			resourceAllocatableNum.EphemeralStorage = uint64(nod.Status.Allocatable.StorageEphemeral().Value())

		}
	}
	return resourceAllocatableNum
}

func GetEachNodeResource(podLister v1.PodLister,nodeLister v1.NodeLister,
	NodeUsedResourceMap NodeUsedResourceMap,
	NodeAllocateResourceMap NodeAllocateResourceMap)(NodeUsedResourceMap,NodeAllocateResourceMap) {

	podList, err := podLister.List(labels.Everything())
	if err != nil {
		log.Println(err)
		panic(err.Error())
	}
	nodeList, err := nodeLister.List(labels.Everything())
	if err != nil {
		log.Println(err)
		panic(err.Error())
	}

	for key, val := range NodeUsedResourceMap {
		currentNodePodsResourceSum := getEachNodePodsResourceRequest(podList, key)
		val.MilliCpu =  currentNodePodsResourceSum.MilliCpu
		val.Memory = currentNodePodsResourceSum.Memory
		NodeUsedResourceMap[key] = NodeUsedResource{val.MilliCpu, val.Memory/1024/1024}
		//log.Println(NodeUsedResourceMap)
	}

	for key, val1 := range NodeAllocateResourceMap {
		currentNodeAllocatableResource := getEachNodeAllocatableNum(nodeList, key)
		val1.MilliCpu = currentNodeAllocatableResource.MilliCpu
		val1.Memory = currentNodeAllocatableResource.Memory
		NodeAllocateResourceMap[key] = NodeAllocateResource{val1.MilliCpu, val1.Memory/1024/1024}
		//log.Println(NodeAllocateResourceMap)
	}
	//log.Println(NodeUsedResourceMap)
	//log.Println(NodeAllocateResourceMap)
	return NodeUsedResourceMap, NodeAllocateResourceMap
}


//Initialize nodeAllocateResourceMap
func initNodeAllocateResourceMap(resourceMap NodeAllocateResourceMap, clusterMasterIp string) NodeAllocateResourceMap {
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
			nodeAllocatedResourceKey := nodeIpPrefix + strconv.Itoa( nodeIpFourthField+i )
			resourceMap[nodeAllocatedResourceKey] = NodeAllocateResource{0, 0}
		}else {
			nodeIpFourthField = nodeIpFourthField + i - 256
			nodeIpThirdField = nodeIpThirdField + 1
			nodeIpPrefix = splitName[0] + "." +splitName[1] + "." + strconv.Itoa(nodeIpThirdField) + "."
			nodeAllocatedResourceKey := nodeIpPrefix + strconv.Itoa(nodeIpFourthField+i)
			resourceMap[nodeAllocatedResourceKey] = NodeAllocateResource{0, 0}
		}
	}
	log.Println(resourceMap)
	return resourceMap
}
//Initialize nodeUsedResourceMap
func initNodeUsedResourceMap(resourceMap NodeUsedResourceMap, clusterMasterIp string) NodeUsedResourceMap {

	//nodeIpPrefix := "192.168.6."
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
			nodeUsedResourceKey := nodeIpPrefix + strconv.Itoa( nodeIpFourthField+i )
			resourceMap[nodeUsedResourceKey] = NodeUsedResource{0, 0}
		}else {
			nodeIpFourthField = nodeIpFourthField + i - 256
			nodeIpThirdField = nodeIpThirdField + 1
			nodeIpPrefix = splitName[0] + "." +splitName[1] + "." + strconv.Itoa(nodeIpThirdField) + "."
			nodeUsedResourceKey := nodeIpPrefix + strconv.Itoa(nodeIpFourthField+i)
			resourceMap[nodeUsedResourceKey] = NodeUsedResource{0, 0}
		}
	}
	log.Println(resourceMap)
	return resourceMap
}

func gatherResource(waiter *sync.WaitGroup,allocateResourceMap NodeAllocateResourceMap,
	usedResourceMap NodeUsedResourceMap, interTimeVal uint32) {
	defer waiter.Done()

	limit := make(chan string,1)
	for{
		clusterAllocatedCpu = 0
		clusterAllocatedMemory = 0
		clusterUsedCpu = 0
		clusterUsedMemory = 0
		limit <- "s"
		time.AfterFunc(time.Duration(interTimeVal)*time.Millisecond, func() {
			//Obtain resource map of Allocatable and Used for each node
			nodeUsedMap, nodeAllocateMap := GetEachNodeResource(podLister, nodeLister,
				usedResourceMap,allocateResourceMap)
			//Traverse nodeAllocateMap
			for _, allocatedVal := range nodeAllocateMap{
				clusterAllocatedCpu += allocatedVal.MilliCpu
				clusterAllocatedMemory += allocatedVal.Memory
			}
			for _, usedVal := range nodeUsedMap{
				clusterUsedCpu += usedVal.MilliCpu
				clusterUsedMemory += usedVal.Memory
			}
			log.Println("****************************************************")
			log.Printf("Current time:%v\n",time.Now().UnixNano()/1e6)
			log.Printf("clusterAllocatedCpu = %d, clusterUsedCpu = %d\n",clusterAllocatedCpu,clusterUsedCpu)
			log.Printf("clusterAllocatedMem = %d, clusterUsedMem = %d\n",clusterAllocatedMemory,clusterUsedMemory)
			log.Println("****************************************************")
			<- limit
		})
	}
}

func main() {
	//Store log
	logFile, err := os.OpenFile("/home/usage.txt", os.O_CREATE | os.O_APPEND | os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout,logFile)
	log.SetOutput(mw)
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

	nodeAllocateResourceMap := make(NodeAllocateResourceMap)
	nodeUsedResourceMap := make(NodeUsedResourceMap)
	allocateResourceMap := initNodeAllocateResourceMap(nodeAllocateResourceMap, masterIp)
	usedResourceMap := initNodeUsedResourceMap(nodeUsedResourceMap, masterIp)

	//Create chan for informer
	stopper := make(chan struct{})
	defer close(stopper)
	waiter := sync.WaitGroup{}
	waiter.Add(1)

	//Create K8s's client
	clientset = GetRemoteK8sClient()
	//Create Informer
	podLister, nodeLister, namespaceLister = InitInformer(stopper,"/kubelet.kubeconfig")

	//Gather resource periodically
	go gatherResource(&waiter,allocateResourceMap,usedResourceMap,interval)

	defer runtime.HandleCrash()

	waiter.Wait()
}
