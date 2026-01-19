package simple

import (
	"testing"
)

func BenchmarkFontFace(b *testing.B) {
	t, _ := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t.FontFace()
	}
}
