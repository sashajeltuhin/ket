package main

import (
	"os"

	"github.com/sashajeltuhin/ket/provision/openstack"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision is a tool for making Kubernetes capable infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	SilenceUsage: true,
}

func init() {

	rootCmd.AddCommand(openstack.Cmd())

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
