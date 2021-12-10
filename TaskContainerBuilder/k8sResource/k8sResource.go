package k8sResource

import (
	_ "golang.org/x/net/context"
	_ "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"

	"k8s.io/client-go/tools/clientcmd"
	"log"
	"path/filepath"
)
//testing k8s cluster API by client-go packages
var clientset *kubernetes.Clientset
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
var resourceRequestNum resourceRequest
var resourceAllocatableNum resourceRequest
type ResidualResourceMap map[string]NodeResidualResource

type NodeUsedResourceMap map[string]NodeUsedResource
type NodeAllocateResourceMap map[string]NodeAllocateResource
var readK8sConfigNum uint64
//GetK8sApiResource


func getK8sClient(configfile string) *kubernetes.Clientset {
	//k8sconfig= flag.String("k8sconfig","/opt/kubernetes/cfg/kubelet.kubeconfig","kubernetes config file path")
	//flag.Parse()
	//var k8sconfig string
	//if configfile == "/kubelet.kubeconfig" {
		//k8sconfig, err  := filepath.Abs(filepath.Dir("/opt/kubernetes/cfg/kubelet.kubeconfig"))

	k8sconfig, err  := filepath.Abs(filepath.Dir("/opt/kubernetes/cfg/kubelet.kubeconfig"))
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

	//viper.AddConfigPath("/opt/kubernetes/cfg/")     //设置读取的文件路径
	//viper.SetConfigName("kubelet") //设置读取的文件名
	//viper.SetConfigType("yaml")   //设置文件的类型
	//k8sconfig := viper.ReadInConfig()
	//viper.WatchConfig()
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
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Println("Connect k8s success.")
	}
	return clientset
}

func recoverPodListerFail() {
	if r := recover(); r!= nil {
		log.Println("recovered from ", r)
	}
}
func GetK8sEachNodeResource(podLister v1.PodLister,nodeLister v1.NodeLister,
	ResourceMap ResidualResourceMap) ResidualResourceMap {
	defer recoverPodListerFail()

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
    for key, val := range ResourceMap {
    	currentNodePodsResourceSum := getEachNodePodsResourceRequest(podList, key)
    	currentNodeAllocatableResource := getEachNodeAllocatableNum(nodeList, key)
    	val.MilliCpu = currentNodeAllocatableResource.MilliCpu - currentNodePodsResourceSum.MilliCpu
    	val.Memory = currentNodeAllocatableResource.Memory - currentNodePodsResourceSum.Memory
    	ResourceMap[key] = NodeResidualResource{val.MilliCpu, val.Memory/1024/1024}
	}
    log.Println(ResourceMap)
	return ResourceMap
}

func GetEachNodeResource(podLister v1.PodLister,nodeLister v1.NodeLister,
	NodeUsedResourceMap NodeUsedResourceMap,
	NodeAllocateResourceMap NodeAllocateResourceMap)(NodeUsedResourceMap,NodeAllocateResourceMap) {
	defer recoverPodListerFail()

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
	}

	return NodeUsedResourceMap, NodeAllocateResourceMap
}

func GetK8sApiResource(podLister v1.PodLister,nodeLister v1.NodeLister) resourceRequest {
	defer recoverPodListerFail()

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

	allPodsResourceRequest := getPodsResourceRequest(podList)

	allNodesResourceAllocatable := getNodesAllocatableNum(nodeList)

	readK8sConfigNum++
	log.Println("--------------------------------------------------------")
	log.Printf("get k8s resource in %dth time through informer.\n",readK8sConfigNum)
	return resourceRequest{
		MilliCpu: allNodesResourceAllocatable.MilliCpu - allPodsResourceRequest.MilliCpu,
		Memory:                     (allNodesResourceAllocatable.Memory - allPodsResourceRequest.Memory) / 1024 / 1024,
		EphemeralStorage:           allNodesResourceAllocatable.EphemeralStorage - allPodsResourceRequest.EphemeralStorage,

	}

}


//Traverse all hte pods of Nodes in K8s Cluster and obtain the account of request of all pods
func getPodsResourceRequest(pods []*corev1.Pod) resourceRequest {

	defer recoverPodListerFail()
	resourceRequestNum := resourceRequest{0, 0, 0}
	for _, pod := range pods {
		if pod.Status.HostIP != "192.168.6.109" {
		//if (pod.Status.HostIP != "121.250.173.190")&&(pod.Status.HostIP != "121.250.173.191")&&(pod.Status.HostIP != "121.250.173.192"){
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
				//log.Printf("cpu = %d\n",resourceRequestNum.MilliCpu)
				//log.Printf("mem = %d\n",resourceRequestNum.Memory/1024/1024)
			}
		}
	}
	return resourceRequestNum
}
// obtain the account of allocatable of each node in cluster
func getEachNodePodsResourceRequest(pods []*corev1.Pod,nodeName string) resourceRequest {
	defer recoverPodListerFail()
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

// obtain the account of allocatable of each node in cluster
func getNodesAllocatableNum(nodes []*corev1.Node) resourceRequest {
	resourceAllocatableNum = resourceRequest{0,0,0}
	for _, nod := range nodes {
        //if nod.Name[0:9] != "admiralty" {
        //if nod.Name[0:10] == "k8s-2-node" {
		if nod.Name[0:9] == "k8s4-node" {
			//if (nod.Name != "121.250.173.190")&&(nod.Name != "121.250.173.191")&&(nod.Name != "121.250.173.192"){
			resourceAllocatableNum.MilliCpu += uint64(nod.Status.Allocatable.Cpu().MilliValue())
			resourceAllocatableNum.Memory += uint64(nod.Status.Allocatable.Memory().Value())
			resourceAllocatableNum.EphemeralStorage += uint64(nod.Status.Allocatable.StorageEphemeral().Value())

		}
	}
	return resourceAllocatableNum
}

func getEachNodeAllocatableNum( nodes []*corev1.Node, nodeName string) resourceRequest {
	defer recoverPodListerFail()
	resourceAllocatableNum = resourceRequest{0,0,0}
	for _, nod := range nodes {
		//if nod.Name[0:9] != "admiralty" {
		//if nod.Name[0:12] == "k8s-2-node-1" {
		if nod.Name == nodeName {
				resourceAllocatableNum.MilliCpu = uint64(nod.Status.Allocatable.Cpu().MilliValue())
				resourceAllocatableNum.Memory = uint64(nod.Status.Allocatable.Memory().Value())
				resourceAllocatableNum.EphemeralStorage = uint64(nod.Status.Allocatable.StorageEphemeral().Value())
	    }
	}
	return resourceAllocatableNum
}


