package http

import (
	"testing"

	"github.com/spiral/roadrunner/v2/plugins/http"
)

var samples = []struct {
	in  string
	out []string
}{
	{"key", []string{"key"}},
	{"key[subkey]", []string{"key", "subkey"}},
	{"key[subkey]value", []string{"key", "subkey", "value"}},
	{"key[subkey][value]", []string{"key", "subkey", "value"}},
	{"key[subkey][value][]", []string{"key", "subkey", "value", ""}},
	{"key[subkey] [value][]", []string{"key", "subkey", "value", ""}},
	{"key [ subkey ] [ value ] [ ]", []string{"key", "subkey", "value", ""}},
}

func Test_FetchIndexes(t *testing.T) {
	for i := 0; i < len(samples); i++ {
		r := http.FetchIndexes(samples[i].in)
		if !same(r, samples[i].out) {
			t.Errorf("got %q, want %q", r, samples[i].out)
		}
	}
}

func BenchmarkConfig_FetchIndexes(b *testing.B) {
	for _, tt := range samples {
		for n := 0; n < b.N; n++ {
			r := http.FetchIndexes(tt.in)
			if !same(r, tt.out) {
				b.Fail()
			}
		}
	}
}

func same(in, out []string) bool {
	if len(in) != len(out) {
		return false
	}

	for i, v := range in {
		if v != out[i] {
			return false
		}
	}

	return true
}
