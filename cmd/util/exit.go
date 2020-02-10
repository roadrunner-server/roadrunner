package util

import (
	"os"
)

// ExitWithError prints error and exits with error code`.
func ExitWithError(err error) {
	errP := Panicf("<red+hb>Error:</reset> <red>%s</reset>\n", err)
	if errP != nil {
		// in case of error during Panicf, print this error via build-int print function
		println("error occurred during fmt.Fprint: " + err.Error())
	}
	os.Exit(1)
}
