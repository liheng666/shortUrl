package myconfig

import (
	"os"
	"encoding/json"
	"shortUrl/app/db"
)

type MyConfig struct {
	Db            db.Db  //mysql配置
	ServerAddress string // 服务器监听地址
}

func LoadConfig(path string) MyConfig {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(file)

	conf := MyConfig{}
	err = decoder.Decode(&conf)
	if err != nil {
		panic(err)
	}

	return conf
}
