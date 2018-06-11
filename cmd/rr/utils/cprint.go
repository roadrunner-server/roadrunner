package utils

import (
	"fmt"
	"gopkg.in/AlecAivazis/survey.v1/core"
	"regexp"
	"strings"
)

// Printf works identically to fmt.Print but adds `<white+hb>color formatting support for CLI</reset>`.
func Printf(format string, args ...interface{}) {
	fmt.Print(Sprintf(format, args...))
}

// Sprintf works identically to fmt.Sprintf but adds `<white+hb>color formatting support for CLI</reset>`.
func Sprintf(format string, args ...interface{}) string {
	r, err := regexp.Compile(`<([^>]+)>`)
	if err != nil {
		panic(err)
	}

	format = r.ReplaceAllStringFunc(format, func(s string) string {
		return fmt.Sprintf(`{{color "%s"}}`, strings.Trim(s, "<>/"))
	})

	out, err := core.RunTemplate(fmt.Sprintf(format, args...), nil)
	if err != nil {
		panic(err)
	}

	return out
}
