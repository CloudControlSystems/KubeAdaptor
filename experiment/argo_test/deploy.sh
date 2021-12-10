#!/bin/bash

cd ./deploy
./edit.sh
cd ..
#deploy
kubectl apply -f ./deploy/rbac-argo.yaml
kubectl apply -f ./deploy/resourceUsage.yaml

for((i=0;i<=99;i++))
do
echo ":----------"
echo "i:$i"
    start_time=$(date +%s )
    #echo $start_time
    argo submit ./deploy/Montage.yaml -n argo
    #argo submit ./deploy/Epigenomics.yaml -n argo
    #argo submit ./deploy/LIGO.yaml -n argo
    #argo submit ./deploy/CyberShake.yaml -n argo

    while [ true ]
    do
      state=`argo list -n argo  |grep "montage"| awk '{print $2}'`
      #state=`argo list -n argo  |grep "epigenomics"| awk '{print $2}'`
      #state=`argo list -n argo  |grep "ligo"| awk '{print $2}'`
      #state=`argo list -n argo  |grep "cybershake"| awk '{print $2}'`
      #echo $state
      if [ $state == "Succeeded" ];then
        podName=`argo list -n argo  |grep "montage"| awk '{print $1}'`
        #podName=`argo list -n argo  |grep "epigenomics"| awk '{print $1}'`
        #podName=`argo list -n argo  |grep "ligo"| awk '{print $1}'`
        #podName=`argo list -n argo  |grep "cybershake"| awk '{print $1}'`
        echo ":$podName"

        argo delete $podName -n argo
        while [ true ]
        do
          podState=`argo list -n argo  |grep "montage"| awk '{print $2}'`
          #podState=`argo list -n argo  |grep "epigenomics"| awk '{print $2}'`
          #podState=`argo list -n argo  |grep "ligo"| awk '{print $2}'`
          #podState=`argo list -n argo  |grep "cybershake"| awk '{print $2}'`
          if [ ! $podState ];then
            echo ":delete argo workflow pod successful."
	    
	    #ip=`kubectl get pods -A -o wide  |grep "resource"| awk '{print $8}'`
            #echo $ip
            #scp root@$ip:/home/usage.txt .
	    #mv usage.txt argo-Montage$i.txt
            #echo "copy resourceUsage's log successful."
	    #./deleteLog.sh

            break
          fi
        done
       else	   
	  if [ ! $state ];then
	    echo ":It's over."
	    break
          fi 
       fi
    done
    end_time=$(date +%s )
    #echo $end_time
    echo "workflowcycle:$(($end_time-$start_time))"

done
#find the hosted node with usage.txt, copy usage.txt file to Master node and delete this usage.txt.
ip=`kubectl get pods -A -o wide  |grep "resource"| awk '{print $8}'`
echo ":$ip"
scp root@$ip:/home/usage.txt . 
mv usage.txt argo-Montage$i.txt
#mv usage.txt argo-Epigenomics$i.txt
#mv usage.txt argo-LIGO$i.txt
#mv usage.txt argo-CyberShake$i.txt
echo ":copy resourceUsage's log successful."
./deleteLog.sh


