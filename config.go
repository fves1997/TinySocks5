package Socks5

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	LocalPort  int    `json:"local_port"`
	Password   string `json:"password"`
	Method     string `json:"method"`
}

const configName = "socks.json"

func NewSocksConfig() *Config {
	config := &Config{}
	config.loadConfig()
	return config
}

func (config *Config) saveDefaultConfig() {
	/* 默认配置 */
	config.Server = "127.0.0.1"
	config.ServerPort = 2355
	config.LocalPort = 1081
	config.Password = "password"

	bytes, _ := json.MarshalIndent(config, "", " ")
	err := ioutil.WriteFile(configName, bytes, 0666)
	if err != nil {
		log.Fatal("写入配置文件失败", err)
	}
}

func (config *Config) loadConfig() {
	_, err := os.Stat(configName)
	if os.IsNotExist(err) {
		log.Println("配置文件不存在，生成默认配置文件")
		config.saveDefaultConfig()
		return
	}
	bytes, err := ioutil.ReadFile(configName)
	if err != nil {
		log.Fatal("读取配置文件失败", err)
	}
	err = json.Unmarshal(bytes, config)
	if err != nil {
		log.Fatal("配置文件不是合格的JSON格式", err)
	}
}
