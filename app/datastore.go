package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"shortUrl/app/db"
	"sync"
	"sync/atomic"
	"time"
)

type Worker struct {
	Queue  *MyQueue
	DB     *sql.DB
	wg     sync.WaitGroup
	Closed chan bool
}

// 构建数据存储实例
func NewWorker(mq *MyQueue, DB *sql.DB) *Worker {
	return &Worker{
		Queue:  mq,
		DB:     DB,
		Closed: make(chan bool, 1),
	}
}

// 存储队列中的数据
func (w *Worker) dbStoreServer() {
	for {
		v, err := w.Queue.Pull()
		if err != nil { // 缓存队列已经关闭
			break
		} else if v == nil { // 队列为空
			time.Sleep(200 * time.Millisecond)
			continue
		}

		err = v.Insert(w.DB)
		if err != nil {
			log.Fatal(err)
		}
	}
	w.wg.Done()
}

// 初始处理数据队列的协程池
// n: 协程数量
func (w *Worker) InitWorker(n int) {
	fmt.Println("数据存储协程池启动...")
	w.wg.Add(n)
	for i := 0; i < n; i++ {
		go w.dbStoreServer()
	}
	w.wg.Wait()
	w.Closed <- true
	fmt.Println("数据存储协程池已关闭")
}

/**
待处理数据队列结构体
*/
type MyQueue struct {
	size   int64            // 队列大小
	queue  chan *db.Request // 缓存channel
	closed uint32           // 队列是否关闭 0.正常 1.关闭
	lock   sync.RWMutex     // 读写锁
}

// 获取实例
func NewMyQueue(n int) *MyQueue {
	return &MyQueue{
		queue: make(chan *db.Request, n),
	}
}

// 将数据写入队列
func (q *MyQueue) Push(v *db.Request) (ok bool, err error) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	if q.closed == 1 {
		return false, errors.New("消息队已关闭")
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
func (q *MyQueue) Pull() (*db.Request, error) {
	select {
	case v, ok := <-q.queue:
		if !ok {
			return nil, errors.New("消息队已关闭") // 队列关闭时
		}
		q.size--

		return v, nil
	default:
		return nil, nil // 队列为空时
	}
}

// 获取队列中消息数量
func (q *MyQueue) Size() int64 {
	return q.size
}

// 关闭队列
func (q *MyQueue) Close() bool {
	if atomic.CompareAndSwapUint32(&q.closed, 0, 1) {
		q.lock.Lock()
		close(q.queue)
		q.lock.Unlock()
		return true
	}
	return false
}
