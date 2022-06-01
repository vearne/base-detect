package service

import (
	"github.com/imroc/req/v3"
	"net/http"
)

func init() {
	client := req.C()
	client.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
		// 无需设置MaxIdleConns
		// MaxIdleConns controls the maximum number of idle (keep-alive)
		// connections across all hosts. Zero means no limit.
		// MaxIdleConns 默认是0，0表示不限制
	}
	req.SetDefaultClient(client)
}
