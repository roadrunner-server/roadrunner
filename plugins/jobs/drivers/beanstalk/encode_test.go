package beanstalk

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"testing"

	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner/v2/utils"
)

func BenchmarkEncodeGob(b *testing.B) {
	tb := make([]byte, 1024*10)
	_, err := rand.Read(tb)
	if err != nil {
		b.Fatal(err)
	}

	item := &Item{
		Job:     "/super/test/php/class/loooooong",
		Ident:   "12341234-asdfasdfa-1234234-asdfasdfas",
		Payload: utils.AsString(tb),
		Headers: map[string][]string{"Test": {"test1", "test2"}},
		Options: &Options{
			Priority: 10,
			Pipeline: "test-local-pipe",
			Delay:    10,
			Timeout:  5,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bb := new(bytes.Buffer)
		err := gob.NewEncoder(bb).Encode(item)
		if err != nil {
			b.Fatal(err)
		}
		_ = bb.Bytes()
		bb.Reset()
	}
}

func BenchmarkEncodeJsonIter(b *testing.B) {
	tb := make([]byte, 1024*10)
	_, err := rand.Read(tb)
	if err != nil {
		b.Fatal(err)
	}

	item := &Item{
		Job:     "/super/test/php/class/loooooong",
		Ident:   "12341234-asdfasdfa-1234234-asdfasdfas",
		Payload: utils.AsString(tb),
		Headers: map[string][]string{"Test": {"test1", "test2"}},
		Options: &Options{
			Priority: 10,
			Pipeline: "test-local-pipe",
			Delay:    10,
			Timeout:  5,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bb, err := json.Marshal(item)
		if err != nil {
			b.Fatal(err)
		}
		_ = bb
	}
}
