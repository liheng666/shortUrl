package shortcode

import (
	"testing"
)

func TestEncode(t *testing.T) {
	str, err := Encode(6500000)
	if err != nil {
		t.Error("Encode 方法返回错误！！！")
	}
	if str != "oOWw" {
		t.Fatal("Encode 返回结果不正确!!!")
	}

}

func TestDecode(t *testing.T) {
	id, err := Decode("oOWw")
	if err != nil {
		t.Error("Decode 方法返回错误！！！")
	}
	if id != 6500000 {
		t.Fatal("Decode 返回结果不正确!!!")
	}
}
