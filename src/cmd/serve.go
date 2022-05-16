/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"net/rpc"

	"github.com/arejula27/measurepymemo/pkg/powerstat"
	"github.com/spf13/cobra"
)

type gatherService struct {
}

func new() *gatherService {
	svc := &gatherService{}
	err := rpc.Register(svc)
	if err != nil {
		panic("Cannot register service")
	}
	return svc
}

func (svc *gatherService) measure(flags RootFlags, reply *powerstat.PowerInfo) error {
	fmt.Println("Call done")
	return nil
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "rpc server",
	Long:  `The application crates a rpc server which listens calls, it replies with the metrics gathered.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serve called")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

}
