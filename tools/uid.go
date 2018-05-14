package tools

import (
	"sync"
	"errors"
	"os"
)

var (
	m        sync.Mutex
	closed   bool
	id       uint64 = 0 // 唯一ID
	filename string     // 唯一ID缓存文件path
	uidError = errors.New("唯一ID发号器已关闭")
	Id       []uint64 // 唯一ID数据记录
)

func Newuid(file string) {
	filename = file

	err := Load(&Id, filename)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		id = Id[0]
	}
}

func GetId() (uint64, error) {
	if closed == true {
		return 0, uidError
	}
	m.Lock()
	defer m.Unlock()

	if closed == true {
		return 0, uidError
	}
	id++
	n := id
	return n, nil
}

func Closeuid() {
	m.Lock()
	closed = true
	m.Unlock()

	Id = []uint64{id}
	err := Store(Id, filename)
	if err != nil {
		panic(err)
	}
}
