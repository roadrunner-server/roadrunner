package util

import (
	"fmt"
	"github.com/mgutz/ansi"
	"regexp"
	"strings"
)

var reg *regexp.Regexp

func init() {
	reg, _ = regexp.Compile(`<([^>]+)>`)
}

// Printf works identically to fmt.Print but adds `<white+hb>color formatting support for CLI</reset>`.
func Printf(format string, args ...interface{}) {
	fmt.Print(Sprintf(format, args...))
}

// Sprintf works identically to fmt.Sprintf but adds `<white+hb>color formatting support for CLI</reset>`.
func Sprintf(format string, args ...interface{}) string {
	format = reg.ReplaceAllStringFunc(format, func(s string) string {
		return ansi.ColorCode(strings.Trim(s, "<>/"))
	})

	return fmt.Sprintf(format, args...)
}
