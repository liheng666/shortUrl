package uuid

import "testing"

// 发号器基准测试
func BenchmarkGetID(b *testing.B) {
	New(10, "")
	for i := 0; i < b.N; i++ {
		GetID()
	}
}
