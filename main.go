package main

import (
	"net/http"
	"fmt"
	"os"
	"shortUrl/tools/shortcode"
	"shortUrl/tools"
	"shortUrl/tools/queue"
	"time"
	"database/sql"
	"shortUrl/app/myconfig"
	"shortUrl/app/db"
	"log"
	"os/signal"
	"syscall"
	"shortUrl/app"
	"net/url"
)

var (
	cacheDir          = "./cache/" // 缓存文件
	queueSize  uint32 = 20000      // 队列大小
	tableCount        = 100        // 数据库分表数量
	myQueue    *queue.MyQueue      // 队列实例
	config     myconfig.MyConfig   // 配置
	DB         *sql.DB             // DB是一个数据库（操作）句柄，代表一个具有零到多个底层连接的连接池。它可以安全的被多个go程同时使用。
	worker     *db.Worker          // 保存数据纤程池
)

func init() {
	// 判断缓存文件夹是否存在
	_, err := os.Stat(cacheDir)
	if os.IsNotExist(err) {
		os.Mkdir(cacheDir, 0700) // 当不存在时创建
	}

	// 加载配置文件
	config = myconfig.LoadConfig("./config.json")

	// DB是一个数据库（操作）句柄，代表一个具有零到多个底层连接的连接池。它可以安全的被多个go程同时使用。
	DB = config.Db.Conn()
	// 初始化数据库表单
	db.CreateTables(DB, tableCount)

	// 获取发号器初始uid
	uid := db.GetInitUid(DB, tableCount)
	// 初始化唯一ID发号器
	tools.Newuid(uint64(uid))

	// 初始缓存队列
	myQueue = queue.NewMyQueue(queueSize)
}

func main() {
	// 程序优雅关闭
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		_ = <-c
		//关闭发号器
		tools.Closeuid()
		// 关闭队列
		myQueue.Close()
		// 判断保存数据进程池是否关闭
		v, ok := <-worker.Closed
		if !ok || v != true {
			log.Fatal("保存数据进程池出错")
		}

		fmt.Println("服务器关闭中......")
		os.Exit(0)

	}()

	// 启动保存数据进程池
	worker = db.NewWorker(myQueue, DB)
	go worker.Start(100)

	mux := http.NewServeMux()
	// icon 请求返回404
	mux.Handle("/favicon.ico", http.NotFoundHandler())

	mux.HandleFunc("/", logMiddleware(index))
	// 获取短链接
	mux.HandleFunc("/getShortUrl", getShortUrl)

	server := http.Server{
		Addr:    config.ServerAddress,
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
	urlStr := r.FormValue("url_srt")
	if urlStr == "" {
		fmt.Fprintf(w, "参数不存在")
	}

	// 规范化长链接，去除Scheme参数
	url, err := url.Parse(urlStr)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	urlStr = url.Host + url.RequestURI()

	id, err := tools.GetId()
	if err != nil {
		panic("获取唯一ID错误")
	}

	fmt.Println("uid:", id)

	str, err := shortcode.Encode(id)
	if err != nil {
		panic("获取短链接编码错误")
	}

	ok, err := myQueue.Push(&db.Request{
		Uid:       id,
		Shortcode: str,
		UrlStr:    urlStr,
		Time:      time.Now(),
	})

	if !ok {
		if err == nil { // 队列已满
			fmt.Fprintf(w, "队列已满")
		} else { // 队列关闭
			fmt.Fprintf(w, err.Error())
		}
	} else {
		data := map[string]string{"url": config.BaseUrl + str}
		app.ApiJson(w, 200, "ok", data)
	}

}

// 短域名转跳
func index(w http.ResponseWriter, r *http.Request) {
	str := r.URL.Path
	rs := []rune(str)
	str = string(rs[1:])
	id, err := shortcode.Decode(str)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "not found")
		return
	}

	myRequest := &db.Request{
		Uid: id,
	}
	err = myRequest.Select(DB)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "not found")
		return
	}

	w.Header().Set("Location", "http://"+myRequest.UrlStr)
	w.WriteHeader(302)
}
