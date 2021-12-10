#!/bin/bash
iface=`ip a | grep "state UP" | awk '{print $2}'| awk -F: '{print $1}' | awk 'NR==1{print}'`
ip_master=`ip a | grep -E "inet *.+$iface" | grep -w inet | awk '{print $2}' | awk -F '/' '{print $1}'`

#删除,master_ip.txt中的$ip_master
sed -i "/$ip_master/d" master_ip.txt

#apt install python
#rpm -ivhU expect/*  --nodeps --force > /dev/null
#alien -i expect/*  > /dev/null
#安装tcl8.4.19,expect5.45
cd expect
tar -zxvf tcl8.4.19-src.tar.gz
tar -zxvf expect5.45.tar.gz
cd ./tcl8.4.19/unix 
./configure
make
make install
cd .. && cd .. && cd expect5.45
./configure --with-tclinclude=/root/deploy-work/no_ssh/expect/tcl8.4.19/generic/ --with-tclconfig=/usr/local/lib/
make
make install
cd ..
cd ..

#安装pssh-2.3.1.tar.gz
tar -zxvf pssh-2.3.1.tar.gz > /dev/null
cd ./pssh-2.3.1
python setup.py build > /dev/null
python setup.py install > /dev/null
cd ..

while read line
do
[ -d "$line/" ] && rm -rf $line
done < master_ip.txt

rm -rf /root/.ssh/

ssh-keygen -t rsa -P '' -f /root/.ssh/id_rsa

while read line;do
ip=`echo $line | cut -d " " -f2`
user_name=`echo $line | cut -d " " -f3`
pass_word=`echo $line | cut -d " " -f4`
expect <<EOF
spawn ssh-copy-id -i /root/.ssh/id_rsa.pub $user_name@$ip
expect {
"yes/no" { send "yes\n";exp_continue }
"password" { send "$pass_word\n" }
}
expect eof
EOF
done < master.txt
pssh -h master_ip.txt -l root 'rm -rf /qzh'
pssh -h master_ip.txt -l root 'mkdir /qzh'
if [ -d '/qzh/' ]
then
rm -rf /qzh
fi
mkdir /qzh

pssh -h master_ip.txt -l root 'rm -rf /root/.ssh/'


echo "step 1"
while read line;do
ip=`echo $line | cut -d " " -f2`
user_name=`echo $line | cut -d " " -f3`
pass_word=`echo $line | cut -d " " -f4`
expect <<EOF
spawn ssh-copy-id -i /root/.ssh/id_rsa.pub $user_name@$ip
expect {
"yes/no" { send "yes\n";exp_continue }
"password" { send "$pass_word\n" }
}
expect eof
EOF
done < master.txt

pssh -h master_ip.txt -l root "ssh-keygen -t rsa -P '' -f /root/.ssh/id_rsa"
pslurp -h master_ip.txt /root/.ssh/id_rsa.pub id_rsa.pub


echo "step 2"
cat /root/.ssh/id_rsa.pub > /root/.ssh/authorized_keys
while read line || [[ -n ${line} ]];
do
ip=`echo $line | cut -d " " -f1`
cat $ip/id_rsa.pub >> /root/.ssh/authorized_keys
done < master_ip.txt

echo "step 3"
pssh -l root -h master_ip.txt 'sed -i "$"d /etc/ssh/sshd_config'
pssh -l root -h master_ip.txt 'sed -i "$"d /etc/ssh/sshd_config'
pssh -l root -h master_ip.txt 'sed -i "$"d /etc/ssh/sshd_config'
pssh -l root -h master_ip.txt 'sed -i "\$a\RSAAuthentication yes" /etc/ssh/sshd_config'
pssh -l root -h master_ip.txt 'sed -i "\$a\PubkeyAuthentication yes" /etc/ssh/sshd_config'
pssh -l root -h master_ip.txt 'sed -i "\$a\PermitRootLogin yes" /etc/ssh/sshd_config'
pscp -l root -h master_ip.txt /root/.ssh/authorized_keys /root/.ssh/authorized_keys
pssh -l root -h master_ip.txt 'sudo systemctl restart sshd'
sed -i "$"d /etc/ssh/sshd_config
sed -i "$"d /etc/ssh/sshd_config
sed -i "$"d /etc/ssh/sshd_config
sed -i '$a\RSAAuthentication yes' /etc/ssh/sshd_config
sed -i '$a\PubkeyAuthentication yes' /etc/ssh/sshd_config
sed -i '$a\PermitRootLogin yes' /etc/ssh/sshd_config
sudo systemctl restart sshd

echo -e "No Secret end"
