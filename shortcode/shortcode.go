package shortcode

var (
	// 64进制使用到的字符列表
	strCode = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/")

	//
	SYSTEM uint32 = 64
)

// 编码
func Encode(id uint32) string {
	var data []rune
	for {
		var r rune   // 下标指向的字符
		var k uint32 // 64进制字符数组下标
		if id < SYSTEM {
			k = id - 1
			r = strCode[k]
			data = append(data, r)
			break
		} else {
			k = id % SYSTEM
			r = strCode[k]
			data = append(data, r)

			id = (id - k) / SYSTEM
		}
	}

	return string(data)
}

// 解码
func Decode(string string) uint32 {
	return 1
}
