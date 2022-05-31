package service

import (
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/vearne/base-detect/internal/config"
	"github.com/vearne/base-detect/internal/consts"
	"github.com/vearne/base-detect/internal/model"
	"net/http"
	"time"
)

func StartAgent() {
	r := gin.Default()
	r.POST("/api/v1/httpdetect", func(c *gin.Context) {

		result := model.AgentHttpDetectResp{}
		result.Data.DataSize = -1
		result.Data.HttpCode = -1
		result.Data.TimeCost = -1

		var param model.AgentHttpDetectReq
		err := c.BindJSON(&param)
		if err != nil {
			result.Status = model.RespStatus{Code: consts.AgentECodeParamError}
			c.JSON(http.StatusBadRequest, &result)
			return
		}

		client := req.C().EnableTraceAll()
		client.SetTimeout(time.Second * time.Duration(param.Timeout))
		resp, err := client.R().Get(param.Target)
		if err != nil {
			result.Status = model.RespStatus{Code: consts.AgentECodeTargetError, Message: err.Error()}
			c.JSON(http.StatusOK, &result)
			return
		}

		result.Status.Code = consts.AgentECodeSuccess
		result.Data.HttpCode = resp.StatusCode
		//fmt.Println(resp.TraceInfo().String())
		result.Data.TimeCost = resp.TraceInfo().TotalTime.Seconds()
		bt, _ := resp.ToBytes()
		result.Data.DataSize = len(bt)

		c.JSON(http.StatusOK, &result)
	})
	r.Run(config.GetAgentConfig().Addr)
}
