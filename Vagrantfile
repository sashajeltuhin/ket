boxes = [
    {
        :name => "etcd1",
        :eth1 => "192.168.205.2",
        :mem => "1024",
        :cpu => "1"
    },{
        :name => "master1",
        :eth1 => "192.168.205.3",
        :mem => "1024",
        :cpu => "1"
    },{
        :name => "worker1",
        :eth1 => "192.168.205.4",
        :mem => "1024",
        :cpu => "1"
    },{
        :name => "ingress1",
        :eth1 => "192.168.205.5",
        :mem => "1024",
        :cpu => "1"
    }
]

Vagrant.configure(2) do |config|

  config.vm.box = "ubuntu/xenial64"
  config.ssh.insert_key = false

  # Add the ssh public key to the node
  config.vm.provision "shell" do |s|
    ssh_pub_key = File.readlines("kismatic-cluster.pem.pub").first.strip
    s.inline = <<-SHELL
      mkdir -p /root/.ssh
      echo #{ssh_pub_key} >> /home/vagrant/.ssh/authorized_keys
      echo #{ssh_pub_key} >> /root/.ssh/authorized_keys
    SHELL
  end

  # Turn off shared folders
  config.vm.synced_folder ".", "/vagrant", id: "vagrant-root", disabled: true

  boxes.each do |opts|
    config.vm.define opts[:name] do |config|
      config.vm.hostname = opts[:name]

      config.vm.provider "vmware_fusion" do |v|
        v.vmx["memsize"] = opts[:mem]
        v.vmx["numvcpus"] = opts[:cpu]
      end

      config.vm.provider "virtualbox" do |v|
        v.customize ["modifyvm", :id, "--memory", opts[:mem]]
        v.customize ["modifyvm", :id, "--cpus", opts[:cpu]]
      end

      config.vm.network :private_network, ip: opts[:eth1]
    end
  end
end