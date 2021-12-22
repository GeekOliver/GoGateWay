package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

type BrowserProxy struct {
}

const (
	Header = "x-Forwarded-For"
)

func (b *BrowserProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("received request %s %s %s\n", req.Method, req.Host, req.RemoteAddr)
	//0.连接池
	transport := http.DefaultTransport
	//1.浅拷贝对象，然后新增属性数据
	outReq := new(http.Request)
	*outReq = *req
	if clientIp, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := outReq.Header[Header]; ok {
			clientIp = strings.Join(prior, ", ") + ", " + clientIp
		}
		outReq.Header.Set(Header, clientIp)
	}
	//fmt.Printf("浅拷贝以后的请求 %s, 原始请求 %s", outReq.Header.Get(Header), req.Header.Get(Header))
	//2.请求下游
	res, err := transport.RoundTrip(outReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	//3.把请求内容返回给上游
	defer res.Body.Close()
	for key, value := range res.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}
	rw.WriteHeader(res.StatusCode)
	io.Copy(rw, res.Body)
}

func main() {
	fmt.Println("serve on: 8010")
	http.Handle("/", &BrowserProxy{})
	err := http.ListenAndServe("0.0.0.0:8010", nil)
	if err != nil {
		fmt.Printf("serve on: 8010 error %s", err.Error())
		return
	}

}
