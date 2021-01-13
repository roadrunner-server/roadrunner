package util

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mgutz/ansi"
)

var (
	reg *regexp.Regexp

	// Colorize enables colors support.
	Colorize = true
)

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
		if !Colorize {
			return ""
		}

		return ansi.ColorCode(strings.Trim(s, "<>/"))
	})

	return fmt.Sprintf(format, args...)
}

// Panicf prints `<white+hb>color formatted message to STDERR</reset>`.
func Panicf(format string, args ...interface{}) error {
	_, err := fmt.Fprint(os.Stderr, Sprintf(format, args...))
	if err != nil {
		return err
	}
	return nil
}
