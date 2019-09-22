package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"shortUrl/app"
	"shortUrl/app/db"
	"shortUrl/app/myconfig"
	"shortUrl/app/mylog"
	"shortUrl/tools"
	"syscall"
)

var worker *app.Worker // 保存数据纤程池

func init() {
	queueSize := 10000 // 队列大小
	tableCount := 100  // 数据库分表数量

	// 初始化自定义log
	mylog.InitLog()
	// 加载配置文件
	myconfig.LoadConfig("./config.json")

	// DB是一个数据库（操作）句柄，代表一个具有零到多个底层连接的连接池。它可以安全的被多个go程同时使用。
	db.InitConn(myconfig.MyConfig.Db)
	// 初始化数据库表单
	db.CreateTables(db.MyDB, tableCount)

	// 获取发号器初始uid
	uid := db.GetInitUid(db.MyDB, tableCount)
	// 初始化唯一ID发号器
	tools.Newuid(uint64(uid))

	// 启动保存数据进程池
	worker = app.NewWorker(queueSize, db.MyDB)
	go worker.InitWorker(100)

	// 程序优雅关闭
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		_ = <-c
		exitFunc()
	}()
}

func main() {
	fmt.Println("短链接服务器启动中...")

	mux := http.NewServeMux()
	// icon 请求返回404
	mux.Handle("/favicon.ico", http.NotFoundHandler())

	mux.HandleFunc("/", app.LogMiddleware(app.Index))
	// 获取短链接
	mux.HandleFunc("/getShortUrl", app.GetShortUrl)

	mux.HandleFunc("/index", app.Home)

	server := http.Server{
		Addr:    myconfig.MyConfig.ServerAddress,
		Handler: mux,
	}

	fmt.Println("短链接服务器启动完成")
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

func exitFunc() {
	//关闭发号器
	tools.Closed()
	fmt.Println("发号器关闭!!!")

	app.MyQueue.Close()
	fmt.Println("队列关闭!!!")

	// 判断保存数据进程池是否关闭
	v, ok := <-worker.Closed
	if !ok || v != true {
		mylog.Error.Fatalln("保存数据进程池出错")
	}
	fmt.Println("数据处理完毕，服务器关闭!!!")
	os.Exit(0)
}
