#!/bin/bash
#kubectl get pods -n argo | awk '{print $1}' > pods.txt
argo list -n argo | awk '{print $1}' > pods.txt
sed -i '1d' pods.txt
#sed -i '1d' pods.txt
#sed -i '$d' pods.txt

while read line
do
#etcdctl del /registry/pods/argo/$line
argo delete $line -n argo
done < pods.txt

