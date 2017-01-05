package vagrant

import (
	"bufio"
	"html/template"
	"os"
)

type Vagrant struct {
	Opts           *InfrastructureOpts
	Infrastructure *Infrastructure
	UnescapedLTLT  template.HTML
	UnescapedGTGT  template.HTML
}

func (v *Vagrant) Write(file *os.File) error {

	v.UnescapedGTGT = template.HTML(">>")
	v.UnescapedLTLT = template.HTML("<<")

	template, err := template.New("vagrantfileOverlay").Parse(vagrantfileOverlay)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(file)

	if err = template.Execute(w, &v); err != nil {
		return err
	}

	w.Flush()

	return nil
}

const vagrantfileOverlay = `boxes = [
    {{range $index,$element := .Infrastructure.Nodes}}{{if $index}},{{end}}{
        :name => "{{.Name}}",
        :eth1 => "{{.IP.String}}",
        :mem => "1024",
        :cpu => "1"
    }{{end}}
]

Vagrant.configure(2) do |config|

  config.vm.box = "{{if .Opts.Redhat}}centos/7{{else}}ubuntu/xenial64{{end}}"
  config.ssh.insert_key = false

  # Add the ssh public key to the node
  config.vm.provision "shell" do |s|
    ssh_pub_key = File.readlines("{{.Infrastructure.PublicSSHKeyPath}}").first.strip
    s.inline = {{.UnescapedLTLT}}-SHELL
      mkdir -p /root/.ssh
      echo #{ssh_pub_key} {{.UnescapedGTGT}} /home/vagrant/.ssh/authorized_keys
      echo #{ssh_pub_key} {{.UnescapedGTGT}} /root/.ssh/authorized_keys
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
end`
