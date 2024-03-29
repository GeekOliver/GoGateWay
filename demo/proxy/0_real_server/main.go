package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type RealServer struct {
	Addr string
}

func (r *RealServer) HelloHandler(w http.ResponseWriter, req *http.Request) {
	upath := fmt.Sprintf("http://%s%s\n", r.Addr, req.URL.Path)
	log.Println(upath)
	realIP := fmt.Sprintf("RemoteAddr=%s,X-Forwarded-For=%v,X-Real-IP=%v\n",
		req.RemoteAddr,
		req.Header.Get("X-Forwarded-For"),
		req.Header.Get("X-Real-IP"))
	io.WriteString(w, upath)
	io.WriteString(w, realIP)
}

func (r *RealServer) ErrorHandler(w http.ResponseWriter, req *http.Request) {
	upath := "error handler"
	log.Println(upath)
	w.WriteHeader(500)
	io.WriteString(w, upath)
}

func (r *RealServer) Run() {
	log.Println("starting httpserver at ", r.Addr)
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.HelloHandler)
	mux.HandleFunc("/base/error", r.ErrorHandler)
	server := &http.Server{
		Addr:         r.Addr,
		Handler:      mux,
		WriteTimeout: 3 * time.Second,
	}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()
}

func main() {
	rs1 := &RealServer{Addr: "127.0.0.1:2003"}
	rs1.Run()
	rs2 := &RealServer{Addr: "127.0.0.1:2004"}
	rs2.Run()

	//监听关闭信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
