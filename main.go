package main

import (
	"net/http"
	"fmt"
	"shortUrl/uuid"
	"os"
	"shortUrl/shortcode"
)

var cacheDir = "./cache/"

var local = "127.0.0.1:8888/"

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
	urlStr := r.FormValue("url_srt")
	if urlStr == "" {
		fmt.Fprintf(w, "参数不存在")
	}

	id, err := uuid.GetID()
	if err != nil {
		panic("获取唯一ID错误")
	}

	str, err := shortcode.Encode(id)
	if err != nil {
		panic("获取短链接编码错误")
	}

	fmt.Fprintf(w, local+str)
}

// 短域名转跳
func index(w http.ResponseWriter, r *http.Request) {
	str := r.URL.Path
	rs := []rune(str)
	str = string(rs[1:])

	w.Header().Set("Location","http://llheng.info")
	w.WriteHeader(302)
	//fmt.Fprintf(w, str)
}
