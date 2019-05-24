package util

import (
	"os"
)

// ExitWithError prints error and exits with error code`.
func ExitWithError(err error) {
	Panicf("<red+hb>Error:</reset> <red>%s</reset>\n", err)
	os.Exit(1)
}
