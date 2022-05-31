package config

import (
	"github.com/spf13/viper"
	"log"
)

type ServerConf struct {
	Addr       string   `json:"string" mapstructure:"addr"`
	AgentAddrs []string `json:"agentAddrs" mapstructure:"agent_addrs"`
}

func InitServerConfig() error {
	log.Println("---InitServerConfig---")
	initOnce.Do(func() {
		var cf = ServerConf{}
		viper.Unmarshal(&cf)
		gcf.Store(&cf)
	})
	return nil
}

func GetServerConfig() *ServerConf {
	return gcf.Load().(*ServerConf)
}
