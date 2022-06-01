# base-detect
基调服务

## 启动agent
```
go run main.go agent --config=./configs/agent1.yaml
go run main.go agent --config=./configs/agent2.yaml
go run main.go agent --config=./configs/agent3.yaml
```


## 启动server
```
go run main.go server --config=./configs/config.server.yaml
```

## 压测
agent
```
wrk -t4 -c200 -d30s --script=./script/post.lua --latency http://127.0.0.1:19291/api/v1/httpdetect
```
server
```
wrk -t4 -c200 -d30s --script=./script/post.lua --latency http://127.0.0.1:19290/api/v1/httpdetect
```