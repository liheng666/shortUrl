package main

import (
	"net/http"
	"fmt"
	"os"
	"shortUrl/shortcode"
	"shortUrl/tools"
	"shortUrl/queue"
	"time"
	"database/sql"
	"shortUrl/myconfig"
	"shortUrl/db"
	"strconv"
	"log"
	"os/signal"
	"syscall"
)

var cacheDir = "./cache/"    // 缓存文件
var queueSize uint32 = 20000 // 队列大小
var myQueue *queue.MyQueue   // 队列实例
var config myconfig.MyConfig // 配置
var DB *sql.DB               // DB是一个数据库（操作）句柄，代表一个具有零到多个底层连接的连接池。它可以安全的被多个go程同时使用。
var tableCount = 100
// 创建
type myRequest struct {
	uid       uint64
	shortcode string
	urlStr    string
	time      time.Time
}

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

	// 初始化唯一ID发号器
	tools.Newuid(cacheDir + "uidcache")

	// 初始缓存队列
	myQueue = queue.NewMyQueue(queueSize)

}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		_ = <-c
		//关闭发号器
		tools.Closeuid()
		// 关闭队列
		myQueue.Close()
		// 判断队列中是否还有值
		for {
			if myQueue.Size() == 0 {
				break
			}
			time.Sleep(100)
		}
		fmt.Println("服务器关闭中......")
		// 关闭
		time.Sleep(1 * time.Second)

		os.Exit(0)

	}()

	go func() {
		err := dbStoreServer()
		if err != nil {
			log.Fatalln("dbStoreServer", err)
		}
	}()

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

// 存储队列中的数据
func dbStoreServer() error {
	for {
		v, err := myQueue.Pull()
		if err != nil { // 缓存队列已经关闭
			fmt.Println(err)
			return err
		} else if v == nil && err == nil { // 队列为空
			fmt.Println("队列为空")
			time.Sleep(1 * time.Second)
			continue
		}
		fmt.Println("queue pull")
		mr, ok := v.(*myRequest)
		if !ok {
			panic("缓存队列中数据类型不正确")
		}
		sql := fmt.Sprintf("insert into short_%d(uid,shortcode,urlstr,time) values(?,?,?,?)", mr.uid%uint64(tableCount))
		stmt, err := DB.Prepare(sql)
		if err != nil {
			log.Fatal(err)
		}
		_, err = stmt.Exec(mr.uid, mr.shortcode, mr.urlStr, mr.time)
		if err != nil {
			log.Fatal(err)
		}
	}
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

	id, err := tools.GetId()
	if err != nil {
		panic("获取唯一ID错误")
	}

	str, err := shortcode.Encode(id)
	if err != nil {
		panic("获取短链接编码错误")
	}

	ok, err := myQueue.Push(&myRequest{
		uid:       id,
		shortcode: str,
		urlStr:    urlStr,
		time:      time.Now(),
	})

	fmt.Println("queue size: ", myQueue.Size(), id)

	if !ok {
		if err == nil { // 队列已满
			fmt.Fprintf(w, "队列已满")
		} else { // 队列关闭
			fmt.Fprintf(w, err.Error())
		}
	} else {
		fmt.Fprintf(w, "llheng.info/"+str)
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
		return
	}
	w.Header().Set("Location", "http://llheng.info/"+strconv.Itoa(int(id)))
	w.WriteHeader(302)
}
