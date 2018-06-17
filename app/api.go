package app

import (
	"encoding/json"
	"net/http"
	"fmt"
)

type BaseApi struct {
	Status  int         `json:"status"`  // 状态值
	Message string      `json:"message"` // 状态信息
	Data    interface{} `json:"data"`    // 返回的信息
}

// http响应返回json
func ApiJson(w http.ResponseWriter, status int, message string, data interface{}) {
	value := BaseApi{
		Status:  status,
		Message: message,
		Data:    data,
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "Application/json")
	fmt.Fprintf(w, string(jsonData[:]))
}
