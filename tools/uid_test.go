package tools

import "testing"

func BenchmarkGetId(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetId()
	}
}
