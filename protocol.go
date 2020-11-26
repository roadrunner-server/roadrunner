package roadrunner

import (
	"os"

	j "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/goridge/v2"
)

var json = j.ConfigCompatibleWithStandardLibrary

type stopCommand struct {
	Stop bool `json:"stop"`
}

type pidCommand struct {
	Pid int `json:"pid"`
}

func sendControl(rl goridge.Relay, v interface{}) error {
	const op = errors.Op("send control")
	if data, ok := v.([]byte); ok {
		err := rl.Send(data, goridge.PayloadControl|goridge.PayloadRaw)
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return errors.E(op, errors.Errorf("invalid payload: %s", err))
	}

	return rl.Send(data, goridge.PayloadControl)
}

func fetchPID(rl goridge.Relay) (int64, error) {
	const op = errors.Op("fetchPID")
	err := sendControl(rl, pidCommand{Pid: os.Getpid()})
	if err != nil {
		return 0, errors.E(op, err)
	}

	body, p, err := rl.Receive()
	if err != nil {
		return 0, errors.E(op, err)
	}
	if !p.HasFlag(goridge.PayloadControl) {
		return 0, errors.E(op, errors.Str("unexpected response, header is missing"))
	}

	link := &pidCommand{}
	err = json.Unmarshal(body, link)
	if err != nil {
		return 0, errors.E(op, err)
	}

	return int64(link.Pid), nil
}
