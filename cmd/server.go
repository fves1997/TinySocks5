package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	. "socks5"
)

func main() {

	sign := make(chan os.Signal)
	signal.Notify(sign, os.Interrupt)
	signal.Notify(sign, os.Kill)
	var config = NewSocksConfig()
	log.Printf("[Server] %s:%d\n", config.Host, config.Port)
	serverSocks := NewServerSocks()
	serverSocks.Listener(fmt.Sprintf("%s:%d", config.Host, config.Port))
	// 内存分析
	go func() {
		serve := http.ListenAndServe(":8888", nil)
		if serve != nil {
			panic(serve)
		}
	}()

	<-sign
}
