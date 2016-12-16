package packet

import (
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func createMinikubeCmd() *cobra.Command {
	opts := &packetOpts{}
	cmd := &cobra.Command{
		Use:   "create-minikube",
		Short: "Creates infrastructure for a single node cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateMinikube(opts)
		},
	}
	cmd.Flags().BoolVar(&opts.CentOS, "useCentos", false, "If present, will install CentOS 7 rather than Ubuntu 16.04")
	cmd.Flags().BoolVarP(&opts.NoPlan, "noplan", "n", false, "If present, foregoes generating a plan file in this directory referencing the newly created nodes")
	cmd.Flags().StringVar(&opts.Region, "region", "us-east", "The region to be used for provisioning machines. One of us-east|us-west|eu-west")
	return cmd
}

func runCreateMinikube(opts *packetOpts) error {
	startTime := time.Now()
	c, err := newFromEnv()
	if err != nil {
		return err
	}

	distro := Ubuntu1604LTS
	if opts.CentOS {
		distro = CentOS7
	}
	provTime := strconv.FormatInt(time.Now().Unix(), 10)
	region, err := regionFromString(opts.Region)
	if err != nil {
		return err
	}

	fmt.Println("Provisioning node")
	hostname := fmt.Sprintf("kismatic-node-%s", provTime)
	nodeID, err := c.CreateNode(hostname, distro, region)
	if err != nil {
		return err
	}

	fmt.Println("Waiting for node to be accessible via SSH. This takes a while...")
	node, err := c.GetSSHAccessibleNode(nodeID, 15*time.Minute, c.SSHKey)
	if err != nil {
		return fmt.Errorf("error waiting for node to be ready")
	}

	fmt.Println()
	fmt.Printf("Finished provisioning nodes on Packet.net in %s\n", time.Now().Sub(startTime))

	if opts.NoPlan {
		fmt.Println("")
		fmt.Printf("%+v", node)
		return nil
	}

	// Write the plan file out
	plan := plan{
		Etcd:                []Node{*node},
		Master:              []Node{*node},
		Worker:              []Node{*node},
		MasterNodeFQDN:      node.PublicIPv4,
		MasterNodeShortName: node.PublicIPv4,
		SSHUser:             node.SSHUser,
		SSHKeyFile:          c.SSHKey,
		AdminPassword:       generateAlphaNumericPassword(),
	}
	template, err := template.New("plan").Parse(overlayNetworkPlan)
	if err != nil {
		return err
	}
	f, err := makeUniqueFile(0)
	if err := template.Execute(f, plan); err != nil {
		return err
	}
	fmt.Printf("Wrote kismatic plan file to %s\n", f.Name())
	return nil
}
