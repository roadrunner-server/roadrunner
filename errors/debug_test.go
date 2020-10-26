// +build debug

package errors

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

var errorLines = strings.Split(strings.TrimSpace(`
	.*/errors/debug_test.go:\d+:  github.com/ValeryPiashchynski/errors.func1:
	.*/errors/debug_test.go:\d+:  ...T.func2:
	.*/errors/debug_test.go:\d+:  ...func3:
	.*/errors/debug_test.go:\d+:  ...func4: func2 invoke func3: Network error:
	func4 operation: error in action
`), "\n")

var errorLineREs = make([]*regexp.Regexp, len(errorLines))

func init() {
	for i, s := range errorLines {
		errorLineREs[i] = regexp.MustCompile(fmt.Sprintf("^%s", s))
	}
}

func TestsDebug(t *testing.T) {
	got := printErr(t, func1())
	lines := strings.Split(got, "\n")
	for i, re := range errorLineREs {
		if i >= len(lines) {
			// Handled by line number check.
			break
		}
		if !re.MatchString(lines[i]) {
			t.Errorf("error does not match at line %v, got:\n\t%q\nwant:\n\t%q", i, lines[i], re)
		}
	}
	if got, want := len(lines), len(errorLines); got != want {
		t.Errorf("got %v lines of errors, want %v", got, want)
	}
}

type T struct{}

func printErr(t *testing.T, err error) string {
	return err.Error()
}

func func1() error {
	var t T
	return t.func2()
}

func (T) func2() error {
	o := Op("func2 invoke func3")
	return E(o, func3())
}

func func3() error {
	return func4()
}

func func4() error {
	o := Op("func4 operation")
	return E(o, Network, Str("error in action"))
}

///Users/0xdev/Projects/repo/errors/debug_test.go:53: github.com/ValeryPiashchynski/errors.func1:
///Users/0xdev/Projects/repo/errors/debug_test.go:58: ...T.func2:
///Users/0xdev/Projects/repo/errors/debug_test.go:62: ...func3:
///Users/0xdev/Projects/repo/errors/debug_test.go:67: ...func4: func2 invoke func3: Network error:
//func4 operation: error in action
