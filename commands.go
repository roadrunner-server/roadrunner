package roadrunner

import (
	"encoding/json"
	"github.com/spiral/goridge"
)

// TerminateCommand must stop underlying process.
type TerminateCommand struct {
	Terminate bool `json:"terminate"`
}

// PidCommand send greeting message between processes in json format.
type PidCommand struct {
	Pid    int `json:"pid"`
	Parent int `json:"parent,omitempty"`
}

// sends control message via relay using JSON encoding
func sendCommand(rl goridge.Relay, command interface{}) error {
	bin, err := json.Marshal(command)
	if err != nil {
		return err
	}

	return rl.Send(bin, goridge.PayloadControl)
}