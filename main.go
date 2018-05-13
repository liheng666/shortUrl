package main

import (
	"net/http"
	"fmt"
	"sync"
	"shortUrl/uuid"
	"time"
	"os"
)

var cacheDir = "./cache/"

func init() {
	// 判断缓存文件夹是否存在
	_, err := os.Stat(cacheDir)
	if os.IsNotExist(err) {
		os.Mkdir(cacheDir, 0700) // 当不存在时创建
	}

	// 初始化唯一ID发号器 步长为10
	uuid.New(10, cacheDir+"uniqueidchdata")
}

func main() {
	//uniqueid_test()

	return

	mux := http.NewServeMux()

	// icon 请求返回404
	mux.Handle("/favicon.ico", http.NotFoundHandler())

	mux.HandleFunc("/", logMiddleware(index))
	// 获取短链接
	mux.HandleFunc("/getShortUrl", getShortUrl)

	server := http.Server{
		Addr:    "127.0.0.1:8888",
		Handler: mux,
	}

	server.ListenAndServe()
}

// log中间件
func logMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 记录需要的信息
		fmt.Println("this is log")

		h(w, r)
	}
}

// 获取短域名
func getShortUrl(w http.ResponseWriter, r *http.Request) {
	//
}

// 短域名转跳
func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

// 发号器测试方法
func uniqueid_test() {
	fmt.Println("开始")
	count := 10000
	var wg sync.WaitGroup
	wg.Add(count)
	t1 := time.Now()
	for i := 0; i < count; i++ {
		go func() {
			id, _ := uuid.GetID()
			if id%1000 == 0 {
				fmt.Println(id)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	runTime := time.Since(t1)
	fmt.Println("运行时长：", runTime)

	defer uuid.Close()
}
