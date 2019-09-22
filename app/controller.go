package app

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"shortUrl/app/db"
	"shortUrl/app/myconfig"
	"shortUrl/app/mylog"
	"shortUrl/tools"
	"time"
)

// web页面
func Home(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./index.html")
	_ = t.Execute(w, nil)
}

// 获取短域名
func GetShortUrl(w http.ResponseWriter, r *http.Request) {
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
		mylog.Error.Fatalf("获取唯一ID错误")
	}

	fmt.Println("uid:", id)

	str, err := tools.Encode(id)
	if err != nil {
		mylog.Error.Fatalf("获取短链接编码错误")
	}

	ok, err := MyQueue.Push(&db.Request{
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
		data := map[string]string{"url": myconfig.MyConfig.BaseUrl + str}
		ApiSuccess(w, 200, "ok", data)
	}

}

// 短域名转跳
func Index(w http.ResponseWriter, r *http.Request) {
	str := r.URL.Path
	rs := []rune(str)
	str = string(rs[1:])
	id, err := tools.Decode(str)
	if err != nil {
		t, _ := template.ParseFiles("./index.html")
		_ = t.Execute(w, nil)
		return
	}

	myRequest := &db.Request{
		Uid: id,
	}
	err = myRequest.Select(db.MyDB)
	if err != nil {
		t, _ := template.ParseFiles("./index.html")
		_ = t.Execute(w, nil)
		return
	}

	w.Header().Set("Location", "http://"+myRequest.UrlStr)
	w.WriteHeader(302)
}
