package model

type ServerHttpDetectReq struct {
	Target  string  `json:"target"`
	Timeout float64 `json:"timeout"`
}

type AgentHttpDetectResult struct {
	Agent    string            `json:"agent"`
	AgentOk  bool              `json:"agentOk"`
	TargetOk bool              `json:"targetOk"`
	Result   *HttpDetectResult `json:"result"`
}

type ServerHttpDetectResp struct {
	Status RespStatus              `json:"status"`
	List   []AgentHttpDetectResult `json:"list"`
}
