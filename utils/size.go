package utils

import (
	"strconv"
	"strings"
)

func ParseSize(size string) int64 {
	if len(size) == 0 {
		return 0
	}

	s, err := strconv.Atoi(size[:len(size)-1])
	if err != nil {
		return 0
	}

	switch strings.ToLower(size[len(size)-1:]) {
	case "k", "kb":
		return int64(s * 1024)
	case "m", "mb":
		return int64(s * 1024 * 1024)
	case "g", "gb":
		return int64(s * 1024 * 1024 * 1024)
	}

	return 0
}
