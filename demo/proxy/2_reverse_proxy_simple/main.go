package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var addr = "127.0.0.1:2005"

func main() {
	//转发地址
	rs1 := "http://127.0.0.1:2003/base"
	u1, e1 := url.Parse(rs1)
	if e1 != nil {
		log.Println(e1)
	}
	//这里其实是NewSingleHostReverseProxy将请求内容目标地址进行更改
	proxy := httputil.NewSingleHostReverseProxy(u1)
	log.Println("starting httpserver at ", addr)
	log.Fatal(http.ListenAndServe(addr, proxy))
}
