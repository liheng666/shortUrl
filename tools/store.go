package tools

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
)

// 在本地文件保存二进制数据
func Store(data interface{}, filename string) error {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, buffer.Bytes(), 0600)

	return nil
}

// 加载本地gob 保存的二进制数据
func Load(data interface{}, filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(raw)
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(data)
	return err
}
