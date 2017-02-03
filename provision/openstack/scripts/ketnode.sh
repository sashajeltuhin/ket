#!/bin/bash
rootPass="^^rootPass^^"
nodeName="^^nodeName^^"
dcip="^^dcip^^"
domainName="^^domainName^^"
domainSuf="^^domainSuf^^"
nodeType="^^nodeType^^"
webIP="^^webIP^^"
webPort="^^webPort^^"
sed -i \"s/mirrorlist=https/mirrorlist=http/\" /etc/yum.repos.d/epel.repo
yum check-update
yum -y install wget libcgroup cifs-utils nano openssh-clients libcgroup-tools unzip iptables-services net-tools bind bind-utils
service cgconfig start
echo "root:$domainPass" | chpasswd
echo "Updating hosts file"
x=$(hostname -I)
eval ipval=($x)
ip=${ipval[0]}
echo "$ip $serverName" >> /etc/hosts
hostnamectl set-hostname $serverName
echo "Updating sshd to allow root login via ssh"
sed -i 's/#\?\(RSAAuthentication\s*\).*$/\1 yes/' /etc/ssh/sshd_config
sed -i 's/#\?\(PermitRootLogin\s*\).*$/\1 yes/' /etc/ssh/sshd_config
sed -i 's/#\?\(PasswordAuthentication\s*\).*$/\1 yes/' /etc/ssh/sshd_config
service sshd restart
echo "Updating domain info in resolv.conf"
cat > /etc/resolv.conf << EOF
nameserver $dcip
search $domainName.$domainSuf
domain $domainName.$domainSuf
EOF
chattr +i /etc/resolv.conf
echo "Add kismatic user"
useradd -d /home/kismaticuser -m kismaticuser
echo "kismaticuser:$domainPass" | chpasswd
echo "kismaticuser ALL = (root) NOPASSWD:ALL" | tee /etc/sudoers.d/kismaticuser
chmod 0440 /etc/sudoers.d/kismaticuser
wget http://$webIP:$webPort/nodeup?type=$nodeType&ip=$ip&name=$nodeName -o /tmp/appscale.log
