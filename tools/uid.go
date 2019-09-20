package tools

import (
	"sync"
	"errors"
	"fmt"
)

var (
	m        sync.Mutex
	closed   bool
	id       uint64 // 唯一ID
	uidError = errors.New("唯一ID发号器已关闭")
)

/**
初始发号器
 */
func Newuid(uid uint64) {
	fmt.Println("发号器初始化...")
	id = uid
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
func Closed() {
	m.Lock()
	closed = true
	m.Unlock()
}
