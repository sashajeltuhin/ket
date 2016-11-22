package packet

import "github.com/spf13/cobra"

type packetOpts struct {
	EtcdNodeCount   uint16
	MasterNodeCount uint16
	WorkerNodeCount uint16
	CentOS          bool
	NoPlan          bool
	Region          string
}

// Cmd returns the command for managing Packet infrastructure
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packet",
		Short: "Provision infrastructure on Packet.net",
	}
	cmd.AddCommand(createCmd())
	cmd.AddCommand(createMinikubeCmd())
	cmd.AddCommand(deleteCmd())
	return cmd
}
