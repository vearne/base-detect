package model

type RespStatus struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
}

type HttpDetectResult struct {
	HttpCode int     `json:"httpCode"`
	TimeCost float64 `json:"timeCost"`
	DataSize float64 `json:"dataSize"`
}
