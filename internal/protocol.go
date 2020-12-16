package internal

import (
	"os"

	j "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3"
)

var json = j.ConfigCompatibleWithStandardLibrary

type StopCommand struct {
	Stop bool `json:"stop"`
}

type pidCommand struct {
	Pid int `json:"pid"`
}

func SendControl(rl goridge.Relay, v interface{}) error {
	const op = errors.Op("send control frame")
	frame := goridge.NewFrame()
	frame.WriteVersion(goridge.VERSION_1)
	frame.WriteFlags(goridge.CONTROL)

	if data, ok := v.([]byte); ok {
		// check if payload no more that 4Gb
		if uint32(len(data)) > ^uint32(0) {
			return errors.E(op, errors.Str("payload is more that 4gb"))
		}

		frame.WritePayloadLen(uint32(len(data)))
		frame.WritePayload(data)
		frame.WriteCRC()

		err := rl.Send(frame)
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return errors.E(op, errors.Errorf("invalid payload: %s", err))
	}

	frame.WritePayloadLen(uint32(len(data)))
	frame.WritePayload(data)
	frame.WriteCRC()

	err = rl.Send(frame)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func FetchPID(rl goridge.Relay) (int64, error) {
	const op = errors.Op("fetchPID")
	err := SendControl(rl, pidCommand{Pid: os.Getpid()})
	if err != nil {
		return 0, errors.E(op, err)
	}

	frameR := goridge.NewFrame()
	err = rl.Receive(frameR)
	if !frameR.VerifyCRC() {
		return 0, errors.E(op, errors.Str("CRC mismatch"))
	}
	if err != nil {
		return 0, errors.E(op, err)
	}
	if frameR == nil {
		return 0, errors.E(op, errors.Str("nil frame received"))
	}

	flags := frameR.ReadFlags()

	if flags&(byte(goridge.CONTROL)) == 0 {
		return 0, errors.E(op, errors.Str("unexpected response, header is missing, no CONTROL flag"))
	}

	link := &pidCommand{}
	err = json.Unmarshal(frameR.Payload(), link)
	if err != nil {
		return 0, errors.E(op, err)
	}

	return int64(link.Pid), nil
}
