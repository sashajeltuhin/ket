package openstack

import (
	"fmt"

	"github.com/spf13/cobra"
)

type KetOpts struct {
	EtcdNodeCount   uint16
	MasterNodeCount uint16
	WorkerNodeCount uint16
	LeaveArtifacts  bool
	RunKismatic     bool
	NoPlan          bool
	ForceProvision  bool
	KeyPairName     string
	InstanceType    string
	OS              string
	Storage         bool
	AdminPass       string
	SSHUser         string
	SSHFile         string
	Domain          string
	Suffix          string
	DNSip           string
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openstack",
		Short: "Provision infrastructure on Openstack.",
		Long: `Provision infrastructure on Openstack.
		
In addition to the commands below, AWS relies on some environment variables and conventions:
Required:
  AWS_ACCESS_KEY_ID: [Required] Your AWS access key, required for all operations
  AWS_SECRET_ACCESS_KEY: [Required] Your AWS secret key, required for all operations

Conditional: (These may be omitted if the -f flag is used)
  AWS_SUBNET_ID: The ID of a subnet to try to place machines into. If this environment variable exists, 
                 it must be a real subnet in the us-east-1 region or all commands will fail.
  AWS_SECURITY_GROUP_ID: The ID of a security group to place all new machines in. Must be a part of the 
                         above subnet or commands will fail.
  AWS_KEY_NAME: The name of a Keypair in AWS to be used to create machines. If empty, we will attempt 
                to use a key named 'kismatic-integration-testing' and fail if it does not exist.
  AWS_SSH_KEY_PATH: The absolute path to the private key associated with the Key Name above. If left blank,
                    we will attempt to use a key named 'kismaticuser.key' in the same directory as the 
					provision tool. This key is important as part of provisioning is ensuring that your
					instance is online and is able to be reached via SSH.
`,
	}

	cmd.AddCommand(CreateCmd())

	return cmd
}

func CreateCmd() *cobra.Command {
	opts := KetOpts{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates infrastructure for a new cluster. For now, only the US East region is supported.",
		Long: `Creates infrastructure for a new cluster. 
		

Smallish instances will be created with public IP addresses. The command will not return until the instances are all online and accessible via SSH.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return makeInfra(opts)
		},
	}

	cmd.Flags().Uint16VarP(&opts.EtcdNodeCount, "etcdNodeCount", "e", 1, "Count of etcd nodes to produce.")
	cmd.Flags().Uint16VarP(&opts.MasterNodeCount, "masterdNodeCount", "m", 1, "Count of master nodes to produce.")
	cmd.Flags().Uint16VarP(&opts.WorkerNodeCount, "workerNodeCount", "w", 1, "Count of worker nodes to produce.")
	cmd.Flags().BoolVarP(&opts.NoPlan, "noplan", "n", false, "If present, foregoes generating a plan file in this directory referencing the newly created nodes")
	cmd.Flags().BoolVarP(&opts.ForceProvision, "force-provision", "f", false, "If present, generate anything needed to build a cluster including VPCs, keypairs, routes, subnets, & a very insecure security group.")
	cmd.Flags().StringVarP(&opts.InstanceType, "instance-type-blueprint", "i", "small", "A blueprint of instance type(s). Current options: micro (all t2 micros), small (t2 micros, workers are t2.medium), beefy (M4.large and xlarge)")
	cmd.Flags().StringVarP(&opts.OS, "operating-system", "o", "ubuntu", "Which flavor of Linux to provision. Try ubuntu, centos or rhel.")
	cmd.Flags().BoolVarP(&opts.Storage, "storage-cluster", "s", false, "Create a storage cluster from all Worker nodes.")
	cmd.Flags().StringVarP(&opts.AdminPass, "admin-pass", "ap", "@ppr3nda", "Admin password")
	cmd.Flags().StringVarP(&opts.SSHUser, "ssh-user", "sshu", "kismatic", "SSH User")
	cmd.Flags().StringVarP(&opts.SSHFile, "ssh-file", "sshf", "/ket/kismaticuser.key", "SSH File")
	cmd.Flags().StringVarP(&opts.Domain, "domain", "d", "ket", "Domain name")
	cmd.Flags().StringVarP(&opts.Domain, "suffix", "suf", "local", "Domain suffix")
	cmd.Flags().StringVarP(&opts.DNSip, "dns-ip", "dns", "10.20.50.175", "Domain IP")

	return cmd
}
func makeInfra(opts KetOpts) error {

	fmt.Print("Provisioning")
	var a Auth
	a.Body.Credentials.Password = "@ppr3nda"
	a.Body.Credentials.Username = "sasha"
	a.Body.Tenant = "f4ec4723e8a541d68ef993b47ef75c94"
	var conf Config
	conf.urlauth = "https://api-trial6.client.metacloud.net:5000/"
	conf.apiverauth = "v2.0"
	var server serverData
	conf.urlcomp = "https://api-trial6.client.metacloud.net:8774/"
	conf.apivercomp = "v2"
	server.Server.Name = "ketautoinst"
	server.Server.ImageRef = "177663bc-0c5e-43b3-99d8-7a457ae4f085"
	server.Server.FlavorRef = "f2c96d12-5454-450c-9ae6-177c4d82eaf3"
	var n network
	n.Uuid = "22e1a428-74a3-4fc1-bd5c-41e10b8ff617"
	server.Server.Networks = append(server.Server.Networks, n)
	var sec secgroup
	sec.Name = "ket"
	server.Server.Security_groups = append(server.Server.Security_groups, sec)
	var nodeID, err = buildNode(a, conf, server, opts, "install")

	if err != nil {
		fmt.Print("Error instantiating Openstack client", err)
		return err
	}

	fmt.Printf("server id  %s", nodeID)

	return nil
}
