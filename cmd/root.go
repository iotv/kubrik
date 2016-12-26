package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use: "kubrik",
	Short: "kubrik is a JSON web service for mg4.",
}