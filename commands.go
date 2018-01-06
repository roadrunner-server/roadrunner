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

func sendCommand(rl goridge.Relay, v interface{}) error {
	bin, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return rl.Send(bin, goridge.PayloadControl)
}

func fetchPid(rl goridge.Relay) (pid int, err error) {
	if err := sendCommand(rl, pidCommand{Pid: os.Getpid()}); err != nil {
		return 0, err
	}

	body, p, err := rl.Receive()
	if !p.HasFlag(goridge.PayloadControl) {
		return 0, fmt.Errorf("unexpected response, `control` header is missing")
	}

	link := &pidCommand{}
	//log.Println(string(body))
	if err := json.Unmarshal(body, link); err != nil {
		return 0, err
	}

	return link.Pid, nil
}
