package main

import (
	"fmt"

	"github.com/vitistack/ipam-api/internal/version"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "ipam-cli",
	Short: "Vitistack IPAM CLI",
	Long:  `Command-line interface for interacting with the Vitistack IPAM system.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use `ipam-cli --help` to see available commands.")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Vitistack IPAM CLI version %s \n", version.Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}
