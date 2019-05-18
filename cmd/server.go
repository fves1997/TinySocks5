package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	. "socks5"
	"strconv"
)

func main() {
	var config = NewSocksConfig()
	log.Println("[Server IP  ]:", config.Server)
	log.Println("[Server Port]:", config.ServerPort)
	log.Println("[Local Port ]:", config.LocalPort)
	serverSocks := NewServerSocks(":" + strconv.Itoa(config.ServerPort))

	// 内存分析
	go func() {
		serve := http.ListenAndServe(":8888", nil)
		if serve != nil {
			panic(serve)
		}
	}()

	serverSocks.Listener()
}
