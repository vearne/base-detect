package model

type AgentHttpDetectReq struct {
	Target  string `json:"target"`
	Timeout int    `json:"timeout"`
}

type AgentHttpDetectResp struct {
	Status RespStatus       `json:"status"`
	Data   HttpDetectResult `json:"data"`
}
