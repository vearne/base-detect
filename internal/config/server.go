package config

type ServerConf struct {
	Addr       string   `json:"string"`
	AgentAddrs []string `json:"agentAddrs"`
}
