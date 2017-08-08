#!/bin/bash
rootPass="^^rootPass^^"
nodeName="^^nodeName^^"
dcip="^^dcip^^"
domainName="^^domainName^^"
domainSuf="^^domainSuf^^"
webPort="^^webPort^^"
postData="^^postData^^"
echo $postData
sed -i \"s/mirrorlist=https/mirrorlist=http/\" /etc/yum.repos.d/epel.repo
yum check-update
yum -y install wget libcgroup cifs-utils nano openssh-clients libcgroup-tools unzip iptables-services net-tools bind bind-utils
service cgconfig start
echo "root:$rootPass" | chpasswd
echo "Updating hosts file"
x=$(hostname -I)
eval ipval=($x)
ip=${ipval[0]}
echo "$ip $nodeName" >> /etc/hosts
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
echo "Install and configure Docker"
yum install -y yum-utils device-mapper-persistent-data lvm2
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum makecache fast
yum install docker-ce
touch /etc/docker/daemon.json
cat > /etc/docker/daemon.json << EOF
{
  "storage-driver": "devicemapper"
}
EOF
mkdir /etc/systemd/system/docker.service.d
touch /etc/systemd/system/docker.service.d/docker.conf
cat > /etc/systemd/system/docker.service.d/docker.conf << EOF
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd --exec-opt native.cgroupdriver=systemd
EOF
systemctl daemon-reload
systemctl start docker
systemctl enable docker

echo "Install git"
yum install -y git

echo "Install go"
cd /tmp
curl -LO https://storage.googleapis.com/golang/go1.7.linux-amd64.tar.gz
tar -C /usr/local -xvzf go1.7.linux-amd64.tar.gz
mkdir -p ~/webket/{bin,pkg,src}
touch /etc/profile.d/path.sh
echo "export PATH=$PATH:/usr/local/go/bin" > /etc/profile.d/path.sh
echo 'export GOBIN="$HOME/webket/bin"' >> ~/.bash_profile
echo 'export GOPATH="$HOME/webket/src"' >> ~/.bash_profile
source /etc/profile && source ~/.bash_profile
echo "Get and install KET orchestrator service"
go get github.com/sashajeltuhin/ket/provision/exec/provision-web
cd $GOPATH/src/github.com/sashajeltuhin/ket/provision/exec/provision-web
docker run -d --name ket -p 8013:8013 -v /ket:/ket sashaz/ketpro
echo "Configure KET user and download KET"
useradd -d /home/kismaticuser -m kismaticuser
echo "kismaticuser:$rootPass" | chpasswd
echo "kismaticuser ALL = (root) NOPASSWD:ALL" | tee /etc/sudoers.d/kismaticuser
chmod 0440 /etc/sudoers.d/kismaticuser
curl https://kismatic-packages-rpm.s3-accelerate.amazonaws.com/kismatic.repo -o /etc/yum.repos.d/kismatic.repo
mkdir /ket
chmod -R 777 /ket
cd /ket
ssh-keygen -t rsa -b 4096 -f kismaticuser.key -P ""

wget -q -O- https://github.com/apprenda/kismatic/releases/download/v1.5.0/kismatic-v1.5.0-linux-amd64.tar.gz | tar -zxf-

chmod +x ./kubectl
mv ./kubectl /usr/local/bin/kubectl
#cp generated/kubeconfig -p $HOME/.kube/config

echo -e "server $dcip\nupdate add $nodeName.$domainName.$domainSuf 3600 A $ip\nsend\n" | nsupdate -v
wget https://raw.githubusercontent.com/sashajeltuhin/ket/master/provision/openstack/scripts/deployapp.sh
wget https://raw.githubusercontent.com/sashajeltuhin/ket/master/provision/openstack/scripts/deploy/osrm.yaml
wget https://raw.githubusercontent.com/sashajeltuhin/ket/master/provision/openstack/scripts/deploy/geo-service.json
wget https://raw.githubusercontent.com/sashajeltuhin/ket/master/provision/openstack/scripts/deploy/geo-ingress.yaml
chmod +x ./deployapp.sh
echo "Post to its own web server"
wget http://$ip:$webPort/install?ip=$ip --post-data $postData -o /tmp/appscale.log
