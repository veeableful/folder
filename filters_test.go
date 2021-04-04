package folder

import (
	"testing"
)

var tokens []string

func init() {
	index := New()
	tokens = index.Analyze("hermione oracle paulina")
}

func BenchmarkLowercaseFilter(b *testing.B) {
	for n := 0; n < b.N; n++ {
		LowercaseFilter(tokens)
	}
}

func BenchmarkPunctuationFilter(b *testing.B) {
	for n := 0; n < b.N; n++ {
		PunctuationFilter(tokens)
	}
}

func BenchmarkStopWordFilter(b *testing.B) {
	for n := 0; n < b.N; n++ {
		StopWordFilter(tokens)
	}
}
