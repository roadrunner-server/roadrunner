package roadrunner

import (
	"encoding/json"
	"fmt"
	"github.com/spiral/goridge"
	"os"
)

type stopCommand struct {
	Stop bool `json:"stop"`
}

type pidCommand struct {
	Pid int `json:"pid"`
}

func sendControl(rl goridge.Relay, v interface{}) error {
	if data, ok := v.([]byte); ok {
		return rl.Send(data, goridge.PayloadControl|goridge.PayloadRaw)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("invalid payload: %s", err)
	}

	return rl.Send(data, goridge.PayloadControl)
}

func fetchPID(rl goridge.Relay) (pid int, err error) {
	if err := sendControl(rl, pidCommand{Pid: os.Getpid()}); err != nil {
		return 0, err
	}

	body, p, err := rl.Receive()
	if !p.HasFlag(goridge.PayloadControl) {
		return 0, fmt.Errorf("unexpected response, header is missing")
	}

	link := &pidCommand{}
	if err := json.Unmarshal(body, link); err != nil {
		return 0, err
	}

	return link.Pid, nil
}
