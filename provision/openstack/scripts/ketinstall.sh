#!/bin/bash
rootPass="^^rootPass^^"
nodeName="^^nodeName^^"
webPort="^^webPort^^"
postData="^^postData^^"
sed -i \"s/mirrorlist=https/mirrorlist=http/\" /etc/yum.repos.d/epel.repo
yum check-update
yum -y install wget libcgroup cifs-utils nano openssh-clients libcgroup-tools unzip iptables-services net-tools
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
echo "Installing Docker"
tee /etc/yum.repos.d/docker.repo <<-'EOF'
[dockerrepo]
name=Docker Repository
baseurl=https://yum.dockerproject.org/repo/main/centos/7
enabled=1
gpgcheck=1
gpgkey=https://yum.dockerproject.org/gpg
EOF
yum install -y http://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.11.1-1.el7.centos.x86_64.rpm
mkdir /etc/systemd/system/docker.service.d
touch /etc/systemd/system/docker.service.d/docker.conf
cat > /etc/systemd/system/docker.service.d/docker.conf << EOF
[Service]
ExecStart=
ExecStart=/usr/bin/docker daemon -H fd:// --exec-opt native.cgroupdriver=systemd
EOF
systemctl daemon-reload
systemctl start docker
systemctl enable docker

echo "Install git"
yum install -y git

echo "Install go"
yum -y install golang
mkdir -p /home/golang
echo 'export GOROOT=/usr/lib/golang' >> /etc/profile.d/go.sh
echo 'export GOBIN=$GOROOT/bin' >> /etc/profile.d/go.sh
echo 'export GOPATH=/home/golang' >> /etc/profile.d/go.sh
echo 'export PATH=$PATH:$GOROOT/bin:$GOPATH/bin' > /etc/profile.d/go.sh
echo '# Golang Path' >> ~/.bashrc
echo 'export GOROOT=/usr/lib/golang' >> ~/.bashrc
echo 'export GOBIN=$GOROOT/bin' >> ~/.bashrc
echo 'export GOPATH=/home/golang' >> ~/.bashrc
echo 'export PATH=$PATH:$GOROOT/bin:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc
source /etc/profile
ldconfig
echo "Get and install KET orchestrator service"
go get github.com/sashajeltuhin/ket/provision/exec/provision-web
cd $GOPATH/src/github.com/sashajeltuhin/ket/provision/exec/provision-web
docker build -t sashaz/ketinstall .
docker run -d --name ket -p 8013:8013 sashaz/ketinstall 
echo "Configure KET user and download KET"
useradd -d /home/kismaticuser -m kismaticuser
echo "kismaticuser:$domainPass" | chpasswd
echo "kismaticuser ALL = (root) NOPASSWD:ALL" | tee /etc/sudoers.d/kismaticuser
chmod 0440 /etc/sudoers.d/kismaticuser
curl https://kismatic-packages-rpm.s3-accelerate.amazonaws.com/kismatic.repo -o /etc/yum.repos.d/kismatic.repo
mkdir /ket
chmod -R 777 /ket
cd /ket
curl -L https://kismatic-installer.s3-accelerate.amazonaws.com/latest/kismatic.tar.gz | tar -zx
ssh-keygen -t rsa -b 4096 -f kismaticuser.key -P ""

#ssh-copy-id -i ./kismaticuser.key.pub kismaticuser@

curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
chmod +x ./kubectl
mv ./kubectl /usr/local/bin/kubectl
#cp generated/kubeconfig -p $HOME/.kube/config

echo "Post to its own web server"
wget http://$ip:$webPort/install --postdata $postData -o /tmp/appscale.log
