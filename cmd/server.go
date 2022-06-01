/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vearne/base-detect/internal/config"
	"github.com/vearne/base-detect/internal/service"
	"log"
	"net/http"
	_ "net/http/pprof"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		loadConfig("server")
		config.InitServerConfig()
		addr := config.GetServerConfig().Addr
		agentAddrs := config.GetServerConfig().AgentAddrs
		go func() {
			log.Println(http.ListenAndServe(":19090", nil))
		}()
		log.Println("addr", addr)
		log.Println("agentAddrs", agentAddrs)
		service.StartServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
