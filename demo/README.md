> demo

## 浏览器代理

+ [浏览器代理browserProxy](browserProxy/main.go)
  + 主要是浏览器正向代理，截取请求，设置内容；请求下游服务；获取响应，改写内容；返回响应给上游请求客户端
  + 运行以后，谷歌浏览器：系统-代理-http，端口设置为8010
  + 接下来主要是访问tianya.cn网站
  + 使用完毕以后记得去掉代理

## 反向代理
+ [模拟服务器](proxy/real_server/main.go)
  + 主要启动两个服务，获取请求path和addr,用来测试反向代理

+ [反向代理reverse_proxy_base](proxy/reverse_proxy_base/main.go)
  + 反向代理的核心知识点为：解析代理地址，并更改请求体的协议和主机



```go
//1.解析代理地址，并更改请求体的协议和主机
proxy, err := url.Parse(addr)
r.URL.Scheme = proxy.Scheme
r.URL.Host = proxy.Host
if err != nil {
log.Print(err)
return
}

//2.请求下游
transport := http.DefaultTransport
rsp, err := transport.RoundTrip(r)
if err != nil {
log.Print(err)
return
}

//3.把下游请求内容返回给上游
for key, value := range rsp.Header {
for _, v := range value {
w.Header().Add(key, v)
}
}
defer rsp.Body.Close()
bufio.NewReader(rsp.Body).WriteTo(w)
```

```
127.0.0.1:8008/abc?sdsdsa=11
r.Addr=127.0.0.1:8008
req.URL.Path=/abc
fmt.Println(req.Host)
```

## http代理

### ReverseProxy基础知识

+ 基于官方ReverseProxy(net/http/httputil)实现一个http代理
  + 功能点
   + 更改内容支持
   + 错误信息回调
   + 支持自定义负载均衡
   + url重写
   + 连接池功能
   + 支持websocket服务
   + 支持https代理
  + 示例
  + 源码分析
  + 拓展功能
   + 4种负载轮询类型
   + 拓展中间件支持：限流、权限、熔断、统计

#### 简单代理转发

+ [http代理](proxy/reverse_proxy_simple/main.go)
  + 主要通过官方ReverseProxy实现http转发代理的流程，流程很简单
  + 在启动realserver以后，然后启动本http代理


```golang
//转发地址
rs1 := "http://127.0.0.1:2003/base"
u1, e1 := url.Parse(rs1)
if e1 != nil {
log.Println(e1)
}

proxy := httputil.NewSingleHostReverseProxy(u1)
log.Println("starting httpserver at ", addr)
log.Fatal(http.ListenAndServe(addr, proxy))
```

请求访问为


```
127.0.0.1:2005
```


#### 更改支持内容




 