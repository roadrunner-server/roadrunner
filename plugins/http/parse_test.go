package http

import "testing"

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
	for _, tt := range samples {
		t.Run(tt.in, func(t *testing.T) {
			r := fetchIndexes(tt.in)
			if !same(r, tt.out) {
				t.Errorf("got %q, want %q", r, tt.out)
			}
		})
	}
}

func BenchmarkConfig_FetchIndexes(b *testing.B) {
	for _, tt := range samples {
		for n := 0; n < b.N; n++ {
			r := fetchIndexes(tt.in)
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
