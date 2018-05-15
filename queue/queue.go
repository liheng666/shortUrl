package queue

import (
	"sync"
	"errors"
	"sync/atomic"
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
func (q *myQueue) Push(v interface{}) (ok bool, err error) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	if q.closed == 1 {
		return false, queueClosedError
	}

	select {
	case q.queue <- v:
		q.size++
		ok = true
	default:
		ok = false
	}

	return
}

// 拉取队列信息
func (q *myQueue) Pull() (interface{}, error) {
	select {
	case v, ok := <-q.queue:
		if !ok {
			return nil, queueClosedError // 队列关闭时
		}
		q.size--

		return v, nil
	default:
		return nil, nil // 队列为空时
	}
}

// 获取队列中消息数量
func (q *myQueue) Size() uint32 {
	return q.size
}

// 关闭队列
func (q *myQueue) Close() bool {
	if atomic.CompareAndSwapUint32(&q.closed, 0, 1) {
		q.lock.Lock()
		close(q.queue)
		q.lock.Unlock()
		return true
	}

	return false
}
