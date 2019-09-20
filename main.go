package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"shortUrl/app"
	"shortUrl/app/db"
	"shortUrl/app/myconfig"
	"shortUrl/tools"
	"syscall"
	"time"
)

var (
	queueSize  = 20000           // 队列大小
	tableCount = 100             // 数据库分表数量
	myQueue    *app.MyQueue      // 队列实例
	config     myconfig.MyConfig // 配置
	DB         *sql.DB           // DB是一个数据库（操作）句柄，代表一个具有零到多个底层连接的连接池。它可以安全的被多个go程同时使用。
	worker     *app.Worker       // 保存数据纤程池
)

func init() {
	// 初始化自定义log
	app.InitLog()

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
	myQueue = app.NewMyQueue(queueSize)
	fmt.Println("数据队列初始化")
	// 启动保存数据进程池
	worker = app.NewWorker(myQueue, DB)
	go worker.InitWorker(100)

	// 程序优雅关闭
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		_ = <-c
		exitFunc()
	}()
}

func exitFunc(){
	//关闭发号器
	tools.Closed()
	// 关闭队列
	myQueue.Close()
	// 判断保存数据进程池是否关闭
	v, ok := <-worker.Closed
	if !ok || v != true {
		app.Error.Fatalln("保存数据进程池出错")
	}
	fmt.Println("服务器关闭!!!")
	os.Exit(0)
}

func main() {
	fmt.Println("短链接服务器启动中...")


	mux := http.NewServeMux()
	// icon 请求返回404
	mux.Handle("/favicon.ico", http.NotFoundHandler())

	mux.HandleFunc("/", logMiddleware(index))
	// 获取短链接
	mux.HandleFunc("/getShortUrl", getShortUrl)

	mux.HandleFunc("/index", home)

	server := http.Server{
		Addr:    config.ServerAddress,
		Handler: mux,
	}

	fmt.Println("短链接服务器启动完成")
	server.ListenAndServe()

}

// log中间件
func logMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 记录需要的信息
		fmt.Println("this is mlog")
		h(w, r)
	}
}

// web页面
func home(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./index.html")
	t.Execute(w, nil)
}

// 获取短域名
func getShortUrl(w http.ResponseWriter, r *http.Request) {
	urlStr := r.FormValue("url_str")
	if urlStr == "" {
		fmt.Fprintf(w, "参数不存在")
		return
	}

	regexpStr, err := regexp.Compile("^(http|https)*.*(cn|com|edu|gov|int|mil|net|org|biz|arpa|info).*$")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	if !regexpStr.MatchString(urlStr) {
		fmt.Fprintf(w, "url格式不正确")
		return
	}

	// 规范化长链接，去除Scheme参数
	url, err := url.Parse(urlStr)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	urlStr = url.Host + url.RequestURI()

	id, err := tools.GetId()
	if err != nil {
		app.Error.Fatalf("获取唯一ID错误")
	}

	fmt.Println("uid:", id)

	str, err := tools.Encode(id)
	if err != nil {
		app.Error.Fatalf("获取短链接编码错误")
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
		app.ApiSuccess(w, 200, "ok", data)
	}

}

// 短域名转跳
func index(w http.ResponseWriter, r *http.Request) {
	str := r.URL.Path
	rs := []rune(str)
	str = string(rs[1:])
	id, err := tools.Decode(str)
	if err != nil {
		t, _ := template.ParseFiles("./index.html")
		t.Execute(w, nil)
		return
	}

	myRequest := &db.Request{
		Uid: id,
	}
	err = myRequest.Select(DB)
	if err != nil {
		t, _ := template.ParseFiles("./index.html")
		t.Execute(w, nil)
		return
	}

	w.Header().Set("Location", "http://"+myRequest.UrlStr)
	w.WriteHeader(302)
}
