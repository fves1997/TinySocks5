package main

import (
	"log"
	. "socks5"
	"strconv"
)

func main() {
	var config = NewSocksConfig()
	log.Println("[Server IP  ]:", config.Server)
	log.Println("[Server Port]:", config.ServerPort)
	log.Println("[Local Port ]:", config.LocalPort)
	serverSocks := NewServerSocks(":" + strconv.Itoa(config.ServerPort))
	serverSocks.Listener()
}
