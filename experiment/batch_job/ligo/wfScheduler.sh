kubectl apply -f rbac-deploy.yaml
kubectl apply -f resourceUsage.yaml
sleep 5s

#for((j=1;j<=100;j++))
#do
echo ":----------"
echo "j:$j"
    start_time=$(date +%s )

kubectl create ns ligo
while [ true ]
do
  state=`kubectl get ns | grep "ligo"|awk '{print $2}'`
  if [ $state == "Active" ];then
    echo "create ligo-workflow namespace successful."
    break
  fi
done

kubectl apply -f priority-job0.yaml
while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ $state == "Completed" ];then
    echo "hello world."
    break
  fi
done

kubectl delete -f priority-job0.yaml

while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ ! $state ];then
    echo "delete priority-job0.yaml successful."
    break
  fi
done

kubectl apply -f priority-job1-5.yaml
while [ true ]
do
  accbool=1
  i=0
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  for bo in $state
  do
    #echo $i
    #echo $bo
    if [ $bo == "Completed" ];then
      accbool=$(($accbool && 1))
      #echo $accbool
    else
      accbool=$(($accbool && 0))
      #echo $accbool
    fi
    i=$(($i+1))
  done
  if [ $accbool -eq 1 ];then
    break
  fi
done

kubectl delete -f priority-job1-5.yaml
while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ ! $state ];then
    echo "delete priority-job1-5.yaml successful."
    break
  fi
done

kubectl apply -f priority-job6-10.yaml
while [ true ]
do
  accbool=1
  i=0
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  for bo in $state
  do
    #echo $i
    #echo $bo
    if [ $bo == "Completed" ];then
      accbool=$(($accbool && 1))
      #echo $accbool
    else
      accbool=$(($accbool && 0))
      #echo $accbool
    fi
    i=$(($i+1))
  done
  if [ $accbool -eq 1 ];then
    break
  fi
done

kubectl delete -f priority-job6-10.yaml
while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ ! $state ];then
    echo "delete priority-job6-10.yaml successful."
    break
  fi
done

kubectl apply -f priority-job11.yaml
while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ $state == "Completed" ];then
    echo "hello world."
    break
  fi
done

kubectl delete -f priority-job11.yaml
while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ ! $state ];then
    echo "delete priority-job11.yaml successful."
    break
  fi
done

kubectl apply -f priority-job12-16.yaml
while [ true ]
do
  accbool=1
  i=0
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  for bo in $state
  do
    #echo $i
    #echo $bo
    if [ $bo == "Completed" ];then
      accbool=$(($accbool && 1))
      #echo $accbool
    else
      accbool=$(($accbool && 0))
      #echo $accbool
    fi
    i=$(($i+1))
  done
  if [ $accbool -eq 1 ];then
    break
  fi
done

kubectl delete -f priority-job12-16.yaml
while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ ! $state ];then
    echo "delete priority-job12-16.yaml successful."
    break
  fi
done

kubectl apply -f priority-job17-21.yaml
while [ true ]
do
  accbool=1
  i=0
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  for bo in $state
  do
    #echo $i
    #echo $bo
    if [ $bo == "Completed" ];then
      accbool=$(($accbool && 1))
      #echo $accbool
    else
      accbool=$(($accbool && 0))
      #echo $accbool
    fi
    i=$(($i+1))
  done
  if [ $accbool -eq 1 ];then
    break
  fi
done

kubectl delete -f priority-job17-21.yaml
while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ ! $state ];then
    echo "delete priority-job17-21.yaml successful."
    break
  fi
done

kubectl apply -f priority-job22.yaml
while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ $state == "Completed" ];then
    echo "hello world."
    break
  fi
done

kubectl delete -f priority-job22.yaml
while [ true ]
do
  state=`kubectl get pods -A |grep "task"| awk '{print $4}'`
  if [ ! $state ];then
    echo "delete priority-job22.yaml successful."
    break
  fi
done

kubectl delete ns ligo
while [ true ]
do
  state=`kubectl get ns | grep "workflow"|awk '{print $2}'`
  if [ ! $state ];then
    echo "delete workflow namespace 'ligo' successful."
    break
  fi
done

    end_time=$(date +%s )
    #echo $end_time
    echo "workflowcycle:$(($end_time-$start_time))"

#done

#find the hosted node with usage.txt, copy usage.txt file to Master node and delete this usage.txt.
ip=`kubectl get pods -A -o wide  |grep "resource"| awk '{print $8}'`
echo ":$ip"
scp root@$ip:/home/usage.txt .
mv usage.txt batch-LIGO$j.txt
echo ":copy resourceUsage's log successful."

./deleteLog.sh


kubectl delete -f rbac-deploy.yaml
kubectl delete -f resourceUsage.yaml
echo "LIGO workflow is over."

