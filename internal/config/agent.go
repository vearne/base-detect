package config

import (
	"github.com/spf13/viper"
)

type AgentConf struct {
	Addr string `json:"string" mapstructure:"addr"`
}

func InitAgentConfig() error {
	initOnce.Do(func() {
		var cf = AgentConf{}
		viper.Unmarshal(&cf)
		gcf.Store(&cf)
	})
	return nil
}

func GetAgentConfig() *AgentConf {
	return gcf.Load().(*AgentConf)
}
