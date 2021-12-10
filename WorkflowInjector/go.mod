module WorkflowInjector

go 1.15

require (
	//github.com/coreos/etcd v3.3.13+incompatible
	github.com/fsnotify/fsnotify v1.4.7
	github.com/golang/protobuf v1.4.2
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/spf13/viper v1.7.0 // indirect
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.24.0
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200520001619-278ece378a50 // indirect

)

replace k8s.io/client-go v11.0.0+incompatible => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90 // indirect

replace k8s.io/apimachinery v0.18.3 => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
