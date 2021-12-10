kubectl delete -f ./deploy/rbac-deploy.yaml
kubectl delete -f ./deploy/storageClass-nfs.yaml
kubectl delete -f ./deploy/workflowInjector-Builder.yaml
kubectl delete -f ./deploy/resourceUsage.yaml
kubectl delete ns -A -l "namespace=task"
kubectl delete pv --all
rm -rf /nfsdata/*
#rm -rf /home/*
rm ./deploy/rbac-deploy.yaml
rm ./deploy/storageClass-nfs.yaml
#rm ./deploy/workflowInjector-Builder.yaml
rm ./deploy/resourceUsage.yaml
#obtain the log file of task-container-builder pod and resource-usage pod.
while [ true ]
do
  state=`kubectl get pods -A -o wide  |grep "usage"| awk '{print $4}'`
  echo $state
  if [ $state == "Terminating" ];then
    ip=`kubectl get pods -A -o wide  |grep "usage"| awk '{print $8}'`
    echo $ip
    scp root@$ip:/home/usage.txt .
    echo "copy resource usage log successful."
    break
  else
    break  
  fi
done

rm ./deploy/workflowInjector-Builder.yaml

while [ true ]
do
  state=`kubectl get pods -A -o wide  |grep "builder"| awk '{print $4}'`
  echo $state
  if [ $state == "Terminating" ];then
    ip=`kubectl get pods -A -o wide  |grep "builder"| awk '{print $8}'`
    echo $ip
    scp root@$ip:/home/log.txt .
    scp root@$ip:/home/exp.txt .
    echo "copy task-container-builder log successful."
    break
  else
    break
  fi
done

#delete log.txt and usage.txt in each node
./deleteLog.sh
