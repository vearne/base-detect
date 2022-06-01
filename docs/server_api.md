## server

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
| timeout  |body|int|是| 单位:秒       ||

#### body
```
{
	"target": "http://news.baidu.com/"
	"timeout": 3
}
```

### Response
| HttpCode | status.code | 说明         |备注|
|:---------|:------------|:-----------|:---|
| 200      | E000        | 成功执行       || 
| 400      | E001        | 参数错误       || 
| 500      | E002        | 内部错误       || 
| 200      | E003        | 部分agent不可用 || 
#### body
| 名称      |位置| 参数类型    | 说明         | 备注         |
|:--------|:---|:--------|:---|:-----------|
| list.agent  |body| string  | agent地址 ||
| list.agentOk  |body| bool  | agent是否正常 ||
| list.targetOk  |body| bool  |target是否正常 ||
| list.result.httpCode |body| int     | http结果 ||
| list.result.timeCost |body| float64 | 耗时 ||
| list.result.dataSize |body| int     | 数据大小 ||

```
{
	"status": {
		"code": "E000",
		"msg": ""
	},
	"list": [{
			"agent": "127.0.0.1:19291",
			"agentOk": true,
			"targetOk": true,
			"result": {
				"httpCode": 200,
				"timeCost": 0.137397334,
				"dataSize": 80545
			}
		}
	]
}
```