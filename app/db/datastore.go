package db

import (
	"time"
	"log"
	"shortUrl/tools/queue"
	"database/sql"
	"sync"
	"fmt"
)

type Worker struct {
	Queue  *queue.MyQueue
	DB     *sql.DB
	wg     sync.WaitGroup
	Closed chan bool
}

// 构建数据存储实例
func NewWorker(myQueue *queue.MyQueue, DB *sql.DB) *Worker {
	return &Worker{
		Queue:  myQueue,
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
		} else if v == nil && err == nil { // 队列为空
			time.Sleep(1 * time.Second)
			continue
		}

		mr, ok := v.(*Request)
		if !ok {
			panic("缓存队列中数据类型不正确")
		}
		err = mr.Insert(w.DB)
		if err != nil {
			log.Fatal(err)
		}
	}
	w.wg.Done()
}

// 开始运行
func (w *Worker) Start(n int) {
	fmt.Println("数据存储协程池启动")
	w.wg.Add(n)
	for i := 0; i < n; i++ {
		go w.dbStoreServer()
	}
	w.wg.Wait()
	w.Closed <- true
	fmt.Println("数据存储协程池已关闭")
}
