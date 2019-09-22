package app

import (
	"fmt"
	"net/http"
)
// 中间件


// log中间件
func LogMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 记录需要的信息
		fmt.Println("this is mlog")
		h(w, r)
	}
}
