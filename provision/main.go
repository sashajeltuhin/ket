package main

import (
	"os"

	"github.com/apprenda/kismatic-provision/provision/aws"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision is a tool for making Kubernetes capable infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(aws.Cmd())
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
