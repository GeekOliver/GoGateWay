package main

import (
	"bufio"
	"log"
	"net/http"
	"net/url"
)

var (
	addr = "http://127.0.0.1:2003"
	port = "2002"
)

func handler(w http.ResponseWriter, r *http.Request) {
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
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("start serving on port " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
