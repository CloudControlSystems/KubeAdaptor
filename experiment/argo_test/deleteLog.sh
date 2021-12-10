#!/bin/bash
# Date  : 2021-01-15 20:07:09
# Author: scg
# Email : uzz_scg@163.com
# Func  : 批量登陆Linux主机并执行命令
username="root"
password="bit123"
port="22"
#timeout=3
#远程登录每个Node删除log.txt,usage.txt
cmd0="rm /home/log.txt"
cmd1="rm /home/exp.txt"
cmd2="rm /home/usage.txt "

login(){
    echo ""
    echo "-------------------------------------------------------- "
    echo "username: $username  password: $password  port: $port  timeout=$timeout"
    echo "command: $cmd"
    echo "Remote exec command script"
    echo "--------------------------------------------------------"
    echo ""

    for host in `cat ipNode.txt`;
    do
        result=""
	echo $host
        #result=`sshpass -p "$password" ssh -p $port -o StrictHostKeyChecking=no -o ConnectTimeout=$timeout $username@$host $cmd0 `
        result=`sshpass -p "$password" ssh -p $port  $username@$host $cmd0 `
        result=`sshpass -p "$password" ssh -p $port  $username@$host $cmd1 `
        result=`sshpass -p "$password" ssh -p $port  $username@$host $cmd2 `
        #echo $host >> result.txt
        #echo $result >> result.txt
    done
}
login
#ls
