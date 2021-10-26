package events

import (
	"strings"

	"github.com/spiral/errors"
)

type wildcard struct {
	prefix string
	suffix string
}

func newWildcard(pattern string) (*wildcard, error) {
	// Normalize
	origin := strings.ToLower(pattern)
	i := strings.IndexByte(origin, '*')

	/*
		http.*
		*
		*.WorkerError
	*/
	if i == -1 {
		dotI := strings.IndexByte(pattern, '.')

		if dotI == -1 {
			// http.SuperEvent
			return nil, errors.Str("wrong wildcard, no * or .")
		}

		return &wildcard{origin[0:dotI], origin[dotI+1:]}, nil
	}

	// pref: http.
	// suff: *
	return &wildcard{origin[0:i], origin[i+1:]}, nil
}

func (w wildcard) match(s string) bool {
	s = strings.ToLower(s)
	return len(s) >= len(w.prefix)+len(w.suffix) && strings.HasPrefix(s, w.prefix) && strings.HasSuffix(s, w.suffix)
}
