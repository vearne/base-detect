/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vearne/base-detect/internal/config"
	"github.com/vearne/base-detect/internal/service"
	"log"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use: "agent",
	Run: func(cmd *cobra.Command, args []string) {
		loadConfig("agent")
		config.InitAgentConfig()
		addr := config.GetAgentConfig().Addr
		log.Println("addr", addr)
		service.StartAgent()
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)
}
