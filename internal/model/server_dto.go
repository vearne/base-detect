package model

type ServerHttpDetectReq struct {
	Target  string `json:"target"`
	Timeout int    `json:"timeout"`
}

type AgentHttpDetectResult struct {
	Agent  string           `json:"agent"`
	Result HttpDetectResult `json:"result"`
}

type ServerHttpDetectResp struct {
	Status RespStatus              `json:"status"`
	List   []AgentHttpDetectResult `json:"list"`
}
