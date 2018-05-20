package main

import (
	"os"
	"encoding/json"
)

type Config struct {
	Db Db //mysql配置
	ServerAddress string  // 服务器监听地址
}

func LoadConfig(path string) Config {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(file)

	conf := Config{}
	err = decoder.Decode(&conf)
	if err != nil {
		panic(err)
	}

	return conf
}
