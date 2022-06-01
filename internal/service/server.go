package service

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/vearne/base-detect/internal/config"
	"github.com/vearne/base-detect/internal/consts"
	"github.com/vearne/base-detect/internal/model"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"time"
)

func StartServer() {
	r := gin.Default()
	r.POST("/api/v1/httpdetect", ServerHttpDetect)
	r.Run(config.GetServerConfig().Addr)
}

func ServerHttpDetect(c *gin.Context) {
	result := model.ServerHttpDetectResp{}

	var param model.ServerHttpDetectReq
	err := c.BindJSON(&param)
	if err != nil {
		result.Status = model.RespStatus{Code: consts.ServerECodeParamError}
		c.JSON(http.StatusBadRequest, &result)
		return
	}

	//ctx := context.Background()
	var g errgroup.Group

	agents := config.GetServerConfig().AgentAddrs
	resChn := make(chan model.AgentHttpDetectResult, len(agents))
	for _, addr := range agents {
		agentAddr := addr
		g.Go(func() error {
			var item model.AgentHttpDetectResult
			item.Agent = agentAddr
			item.AgentOk = true
			item.TargetOk = true

			url := fmt.Sprintf("http://%v/api/v1/httpdetect", agentAddr)
			log.Println("url", url)

			ctx, cancel := context.WithTimeout(context.Background(),
				time.Second*time.Duration(param.Timeout)+time.Millisecond*100)
			defer cancel()
			resp, err := req.SetBody(&param).SetContext(ctx).Post(url)
			if err != nil {
				log.Println("err", err)
				item.AgentOk = false
			} else {
				dresp := model.AgentHttpDetectResp{}
				resp.Unmarshal(&dresp)
				if dresp.Status.Code == consts.AgentECodeTargetError {
					item.TargetOk = false
				} else {
					item.Result = dresp.Data
				}
			}

			resChn <- item
			return nil
		})
	}
	g.Wait()
	close(resChn)

	result.Status.Code = consts.ServerECodeSuccess
	result.List = make([]model.AgentHttpDetectResult, 0, len(agents))
	abnormalCount := 0
	for item := range resChn {
		if !item.AgentOk {
			abnormalCount++
		}
		result.List = append(result.List, item)
	}
	if abnormalCount >= len(agents) {
		result.Status.Code = consts.ServerECodeInternalError
		c.JSON(http.StatusInternalServerError, &result)
		return
	} else if abnormalCount > 0 {
		result.Status.Code = consts.ServerECodeAgentError
	}
	c.JSON(http.StatusOK, &result)
}
