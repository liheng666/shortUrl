package tools

import (
	"sync"
	"errors"
	"os"
)

var (
	m        sync.Mutex
	closed   bool
	id       uint64 // 唯一ID
	Id       uint64 // 唯一ID数据记录
	filename string // 唯一ID缓存文件path
	uidError = errors.New("唯一ID发号器已关闭")
)

/**
初始发号器 会检查是否有缓存文件，有的话会加载缓存的数据
file 缓存文件path
 */
func Newuid(file string) {
	filename = file

	err := Load(&Id, filename)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		id = Id
	}
}

// 获取uid
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

// 关闭应用是调用，会保存当前的发号状态
func Closeuid() {
	m.Lock()
	closed = true
	m.Unlock()

	Id = id
	err := Store(Id, filename)
	if err != nil {
		panic(err)
	}
}
