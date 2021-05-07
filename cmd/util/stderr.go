package util

import (
	"fmt"
	"github.com/spiral/roadrunner"
	"strings"
)

// StdErrOutput outputs rr event into given logger and return false if event was not handled.
func StdErrOutput(event int, ctx interface{}) bool {
	// outputs
	switch event {
	case roadrunner.EventStderrOutput:
		for _, line := range strings.Split(string(ctx.([]byte)), "\n") {
			if line == "" {
				continue
			}

			fmt.Println(strings.Trim(line, "\r\n"))
		}

		return true
	}

	return false
}
