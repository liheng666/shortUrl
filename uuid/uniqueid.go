package uuid

import (
	"sync"
	"sync/atomic"
	"errors"
)

var count uint32 // 发号器通道数量和id添加的步长
//var storeID  []int
var myBuffer *buffer

func init() {
	count = 10 // 默认步长 10
	myBuffer = newBuffer()
}

// 唯一ID结构
type uuid struct {
	id uint32
}

func newuuid(id uint32) *uuid {
	return &uuid{
		id: id,
	}
}

func (uuid *uuid) add() {
	atomic.AddUint32(&uuid.id, count) // 将id添加固定步长
}

type buffer struct {
	ch      chan *uuid   // 缓存通道 用来提高发号效率
	closed  uint32       // ID分发是否关闭 0. 未关闭 1. 已关闭
	closing sync.RWMutex // 用来防止 发号器关闭时产生的竞态条件
}

func newBuffer() *buffer {
	ch := make(chan *uuid, count)
	var i uint32
	for i = 0; i < count; i++ {
		ch <- newuuid(i)
	}

	return &buffer{
		ch: ch,
	}
}

func (buffer *buffer) getID() (uint32, error) {
	// 检测发号器是否已关闭
	if buffer.closed == 1 {
		return 0, errors.New("发号器已关闭")
	}

	buffer.closing.RLock()
	defer buffer.closing.RUnlock()
	// 再次检测发号器是否已关闭
	if buffer.closed == 1 {
		return 0, errors.New("发号器已关闭")
	}

	uuid, ok := <-buffer.ch
	if !ok {
		return 0, errors.New("发号器已关闭")
	}
	uuid.add()
	id := uuid.id
	buffer.ch <- uuid // 使用完毕放回 缓存通道

	return id, nil
}

// 获取唯一ID 公开方法
func GetID() (id uint32, err error) {
	id, err = myBuffer.getID()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// close
func Close() {
	if myBuffer.closed == 1 {
		return
	}

	myBuffer.closing.Lock()
	defer myBuffer.closing.Unlock()

	myBuffer.closed = 1
	close(myBuffer.ch)
}
