package queue

import (
	"sync"
	"errors"
)

var queueClosedError = errors.New("消息队已关闭")

type myQueue struct {
	size   uint32           // 队列大小
	queue  chan interface{} // 缓存channel
	closed uint32           // 队列是否关闭 0.正常 1.关闭
	lock   sync.RWMutex     // 读写锁
}

// 获取实例
func NewMyQueue(n uint32) *myQueue {
	return &myQueue{
		queue: make(chan interface{}, n),
	}
}

// 将数据写入队列
func (q *myQueue) push(v interface{}) (ok bool, err error) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	if q.closed == 1 {
		return false, queueClosedError
	}

	select {
	case q.queue <- v:
		ok = true
	default:
		ok = false
	}

	return
}

func (q *myQueue) pull() (interface{}, error) {

}
