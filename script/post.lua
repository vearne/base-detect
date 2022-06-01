-- example HTTP POST script which demonstrates setting the
-- HTTP method, body, and adding a header

wrk.method = "POST"
wrk.body   = '{"target": "http://news.baidu.com/","timeout": 3}'
wrk.headers["Content-Type"] = "application/json"