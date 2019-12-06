package Socks5

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

const configName = "socks.json"

func NewSocksConfig() *Config {
	config := &Config{}
	config.loadConfig()
	return config
}

func (config *Config) saveDefaultConfig() {
	/* 默认配置 */
	config.Host = "0.0.0.0"
	config.Port = 2355

	bytes, _ := json.MarshalIndent(config, "", " ")
	err := ioutil.WriteFile(configName, bytes, 0666)
	if err != nil {
		log.Fatalf("Write Config Error: %s\n", err.Error())
	}
}

func (config *Config) loadConfig() {
	_, err := os.Stat(configName)
	if os.IsNotExist(err) {
		log.Println("Config isn't exist,Initialize the default configuration")
		config.saveDefaultConfig()
		return
	}
	bytes, err := ioutil.ReadFile(configName)
	if err != nil {
		log.Fatalf("Read Config Error: %s\n", err.Error())
	}
	err = json.Unmarshal(bytes, config)
	if err != nil {
		log.Fatalf("Unmarshal config Error: %s\n", err.Error())
	}
}
