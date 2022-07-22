package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/vearne/base-detect/internal/config"
	"github.com/vearne/base-detect/internal/consts"
	"github.com/vearne/base-detect/internal/model"
	"net/http"
	"time"
)

func StartAgent() {
	r := gin.Default()
	r.POST("/api/v1/httpdetect", AgentHttpDetect)
	r.Run(config.GetAgentConfig().Addr)
}

func AgentHttpDetect(c *gin.Context) {

	result := model.AgentHttpDetectResp{}
	var param model.AgentHttpDetectReq
	err := c.BindJSON(&param)
	if err != nil {
		result.Status = model.RespStatus{Code: consts.AgentECodeParamError}
		c.JSON(http.StatusBadRequest, &result)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(param.Timeout))
	defer cancel()

	client := resty.New()
	resp, err := client.R().
		EnableTrace().
		SetContext(ctx).
		Get(param.Target)

	if err != nil {
		result.Status = model.RespStatus{Code: consts.AgentECodeTargetError, Message: err.Error()}
		c.JSON(http.StatusOK, &result)
		return
	}

	result.Status.Code = consts.AgentECodeSuccess
	result.Data = &model.HttpDetectResult{}
	result.Data.HttpCode = resp.StatusCode()
	result.Data.DataSize = len(resp.Body())

	info := resp.Request.TraceInfo()
	result.Data.DNSLookup = info.DNSLookup
	result.Data.ConnTime = info.ConnTime
	result.Data.TCPConnTime = info.TCPConnTime
	result.Data.TLSHandshake = info.TLSHandshake
	result.Data.ServerTime = info.ServerTime
	result.Data.ResponseTime = info.ResponseTime
	result.Data.TotalTime = info.TotalTime
	result.Data.RemoteAddr = info.RemoteAddr.String()

	c.JSON(http.StatusOK, &result)
}
