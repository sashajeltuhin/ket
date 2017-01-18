# Using Mac and Vagrant

## Envionment Setup

**Managing virtual machines is REALLY HARD**. This is why we use Vagrant. However, it's also why Vagrant and Virtualbox are constantly racing each other with features, simplifications, defects and stabilization. It's strongly urged that you update both fairly often and be prepared to back off a version if it doesn't work for your environment.

1. Install a Vagrant compatible virtual-machine provider such as [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
2. Install (Vagrant)[https://www.vagrantup.com/docs/installation/]
3. Open a Terminal session
4. <table><tr><td>`mkdir ~/kismatic` <br/>
   `cd ~/kismatic`</td> 
   <td>*Make a new directory for Kismatic (~/kismatic would work) and make it the working directory*</td></tr></table>
5. <table><tr><td>`wget -O - https://kismatic-installer.s3-accelerate.amazonaws.com/latest-darwin/kismatic.tar.gz | tar -zx`</td> 
   <td> *Download & unpack Kismatic*</td></tr></table>

## Make a new cluster

6. <table><tr><td>`./provision vagrant create-mini`</td><td> *create a single virtual machine instance kubernetes cluster.*</td></tr></table>
7. <table><tr><td>`./kismatic install apply -f kismatic-cluster.yaml`</td><td> *install Kubernetes*</td></tr></table>

## Tear it down when you're done with it

8. <table><tr><td>`vagrant destroy --force`</td><td> *remove any VM instances created by kismatic on this machine.*</td></tr></table>
