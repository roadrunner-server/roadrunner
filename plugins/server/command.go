package server

import (
	"os"
	"regexp"

	"github.com/spiral/errors"
)

// pattern for the path finding
const pattern string = `^\/*([A-z/.:-]+\.(php|sh|ph))$`

func (server *Plugin) scanCommand(cmd []string) error {
	const op = errors.Op("server_command_scan")
	r, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	for i := 0; i < len(cmd); i++ {
		if r.MatchString(cmd[i]) {
			// try to stat
			_, err := os.Stat(cmd[i])
			if err != nil {
				return errors.E(op, errors.FileNotFound, err)
			}

			// stat successful
			return nil
		}
	}
	return errors.E(errors.Str("scan failed, possible path not found"), op)
}
