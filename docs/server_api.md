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
| HttpCode | status.code | 说明        |备注|
|:---------|:------------|:----------|:---|
| 200      | E000        | 成功执行      || 
| 400      | E001        | 参数错误      || 
| 500      | E002        | 内部错误      || 
| 503      | E003        | 有agent不可用 || 
#### body
| 名称      |位置|参数类型| 说明         | 备注               |
|:--------|:---|:---|:---|:-----------------|
| list.agent  |body|string| agent地址 ||
| list.result.httpCode |body|int| http结果 | -1 表示未能获得结果      |
| list.result.timeCost |body|float64| 耗时 | 单位:秒 -1 表示未能获得结果 |
| list.result.dataSize |body|float64| 数据大小 | 单位:KB    -1 表示未能获得结果        |

```
{
	"status": {
		"code": "E000",
		"msg": "success"
	},
	"list": [{
		"agent": "192.168.20.100:7983",
		"result": {
			"httpCode": 200,
			"timeCost": 2.5,
			"dataSize": 25.5
		}
	}]
}
```