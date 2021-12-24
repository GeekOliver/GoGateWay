package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var addr = "127.0.0.1:2005"

func main() {
	//转发地址
	rs1 := "http://127.0.0.1:2002"
	u1, e1 := url.Parse(rs1)
	if e1 != nil {
		log.Println(e1)
	}
	//这里其实是NewSingleHostReverseProxy将请求内容目标地址进行更改
	proxy := NewSingleHostReverseProxy(u1)
	log.Println("starting httpserver at ", addr)
	log.Fatal(http.ListenAndServe(addr, proxy))
}

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
		req.Header.Set("X-Real-IP", req.RemoteAddr)
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

//这里是组装最终请求路径
func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
