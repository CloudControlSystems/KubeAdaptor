kubectl create ns workflow
while [ true ]
do
  state=`kubectl get ns | grep "workflow"|awk '{print $2}'`
  if [ $state == "Active" ];then
    echo "create workflow namespace successful."
    break
  fi
done

kubectl apply -f one-job0.yaml
while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ $state == "Completed" ];then
    echo "hello world."
    break
  fi
done

kubectl delete -f one-job0.yaml

while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ ! $state ];then
    echo "delete one-job0.yaml successful."
    break
  fi
done

echo "one tasks is over."

etcdctl del /registry/namespaces/workflow

while [ true ]
do
  state=`kubectl get ns | grep "workflow"|awk '{print $2}'`
  if [ ! $state ];then
    echo "delete workflow namespace successful."
    break
  fi
done
echo "delete namespace successful."
