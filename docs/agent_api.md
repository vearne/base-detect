## agent

### 1. 接收http检测请求

### Request

#### method
POST

#### path
/api/v1/httpdetect

#### params

| 名称      |位置|参数类型|是否必须| 说明         |备注|
|:--------|:---|:---|:---|:-----------|:---|
| target  |body|string|是| 需要探测的URL地址 ||
| timeout  |body|int64|是| 单位:秒       ||

#### body
```
{
	"target": "http://news.baidu.com/"
	"timeout": 3
}
```

### Response
| HttpCode | status.code | 说明        |备注|
|:---------|:------------|:----------|:---|
| 200      | A000        | 成功执行      || 
| 400      | A001        | 参数错误      || 
| 500      | A002        | 内部错误      || 
| 200      | A003        | 检测目标时，建立连接失败或访问超时      || 


#### body
| 名称      |位置| 参数类型    | 说明           | 备注               |
|:--------|:---|:--------|:-------------|:-----------------|
| data.httpCode |body| int64     | http结果       |   |
| data.dataSize |body| int64     | 数据大小 单位:byte |     |
| data.DNSLookup |body| int64     | DNS解析耗时 | 单位: ns  |
| data.connTime |body| int64     | 获得一个合法的连接所用的耗时 |  单位: ns   |
| data.TCPConnTime |body| int64     | 建立TCP连接的耗时 |  单位: ns   |
| data.TLSHandshake |body| int64     | TLS握手耗时 | 单位: ns    |
| data.serverTime |body| int64     | 请求发出到收到第一个响应字节的耗时 |  基本可以理解为server端的处理时间紧 + 0.5RTT，单位: ns   |
| data.responseTime |body| int64 | 接收响应的耗时 | 从收到第一个响应字节算起，单位: ns    |
| data.totalTime |body| int64  | 总耗时 |  单位: ns   |
| data.remoteAddr |body| string | server端地址 |     |

**注意**: 如果检测失败data 为null
```
{
	"status": {
		"code": "A000",
		"msg": "success"
	},
	"data": {
		"httpCode": 200,
		"dataSize": 209719,
		"DNSLookup": 4710667,
		"connTime": 141359417,
		"TCPConnTime": 12219542,
		"TLSHandshake": 124070375,
		"serverTime": 20789125,
		"responseTime": 1383625,
		"totalTime": 163397042,
		"remoteAddr": "101.42.231.114:443"
	}
}
```