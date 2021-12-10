#!/bin/bash
#get master ip in this K8s cluster
iface=`ip a | grep "state UP" | awk '{print $2}'| awk -F: '{print $1}' | awk 'NR==1{print}'`
ip_master=`ip a | grep -E "inet *.+$iface" | grep -w inet | awk '{print $2}' | awk -F '/' '{print $1}'`

cp rbac-argo.yaml.bak rbac-argo.yaml
#cp storageClass-nfs.yaml.bak storageClass-nfs.yaml
cp resourceUsage.yaml.bak resourceUsage.yaml
cp ./Montage.yaml.bak Montage.yaml
#cp ./Epigenomics.yaml.bak Epigenomics.yaml
#cp ./LIGO.yaml.bak LIGO.yaml
#cp ./CyberShake.yaml.bak CyberShake.yaml
#替换IP

sed -i "s/0.0.0.0/$ip_master/g" resourceUsage.yaml
#sed -i "s/0.0.0.0/$ip_master/g" task-dag.yaml


ip1=`echo $ip_master | awk -F . '{print $1}'`
ip2=`echo $ip_master | awk -F . '{print $2}'`
ip3=`echo $ip_master | awk -F . '{print $3}'`
ip4=`echo $ip_master | awk -F . '{print $4}'`


for((i=0;i<=6;i++))
do
  ip0=$(($ip4+$i))
  if [ $ip0 -lt 256 ]
  then
    ip10=$ip1.$ip2.$ip3.$ip0
    sed -i "s/$i.$i.$i.$i/$ip10/g" rbac-argo.yaml
  else
    ip0=$(($ip0-256))
    ip33=$(($ip3+1))
    ip10=$ip1.$ip2.$ip33.$ip0
    sed -i "s/$i.$i.$i.$i/$ip10/g" rbac-argo.yaml
  fi
done



