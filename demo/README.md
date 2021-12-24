> demo

## 浏览器代理

+ [浏览器代理browserProxy](browserProxy/main.go)
  + 主要是浏览器正向代理，截取请求，设置内容；请求下游服务；获取响应，改写内容；返回响应给上游请求客户端
  + 运行以后，谷歌浏览器：系统-代理-http，端口设置为8010
  + 接下来主要是访问tianya.cn网站
  + 使用完毕以后记得去掉代理

## 反向代理
+ [模拟服务器](proxy/0_real_server/main.go)
  + 主要启动两个服务，获取请求path和addr,用来测试反向代理

+ [反向代理reverse_proxy_base](proxy/1_reverse_proxy_base/main.go)
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

+ [http代理](proxy/2_reverse_proxy_simple/main.go)
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

+ [通过reverseproxy更改内容](proxy/3_reverse_proxy/main.go)

本小结主要通过net/http/httputil/reverseproxy.go提供的接口实现

```go
type ReverseProxy struct {
	//控制器,是一个函数，函数可以修改请求体内容
	Director func(*http.Request)
	//连接池，如果没有申明，则会使用默认的
	Transport http.RoundTripper
	//刷新到客户端的刷新间隔效率
	FlushInterval time.Duration
	//错误记录器
	ErrorLog *log.Logger
	//定义缓冲池，在复制http相应时使用，用以提高请求效率
	BufferPool BufferPool
	//修改请求结果
	ModifyResponse func(*http.Response) error
	//错误处理回调函数，如果为nil，则遇到错误会显示502
	ErrorHandler func(http.ResponseWriter, *http.Request, error)
}
```
 
基于已经有的`NewSingleHostReverseProxy`添加修改返回的响应`modifyResponseFunc`

```go

func NewSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	//请求为"http://127.0.0.1:2003/base?name=oliver"
	//RawQuery: name=oliver
	//Scheme: http
	//path: /base
	//host: 127.0.0.1:2003
	targetQuery := target.RawQuery
	//创建一个director方法，这里修改的是请求前的内容
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		//这里实际请求的地址会在target后面加上原始请求的path
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	//官方没有实现支持修改返回内容，但是预留了接口
	modifyResponseFunc := func(res *http.Response) error {
		//关键内容，将原始的响应取出先，然后修改，最终写回到
		oldPayload, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		//修改返回体
		newPlayload := []byte("hello " + string(oldPayload))
		//将内容写回到body
		res.Body = ioutil.NopCloser(bytes.NewBuffer(newPlayload))
		//修改最终的内容长度
		res.ContentLength = int64(len(newPlayload))
		//同时设置头里面的Content-Length
		res.Header.Set("Content-Length", fmt.Sprint(res.ContentLength))
		return nil
	}

	//将具体这俩步骤，请求体修改，响应修改传入
	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponseFunc,
	}
}
```