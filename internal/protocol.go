package internal

import (
	"os"
	"sync"

	j "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/spiral/goridge/v3/pkg/relay"
)

var json = j.ConfigCompatibleWithStandardLibrary

type StopCommand struct {
	Stop bool `json:"stop"`
}

type pidCommand struct {
	Pid int `json:"pid"`
}

var fPool = sync.Pool{New: func() interface{} {
	return frame.NewFrame()
}}

func getFrame() *frame.Frame {
	return fPool.Get().(*frame.Frame)
}

func putFrame(f *frame.Frame) {
	f.Reset()
	fPool.Put(f)
}

func SendControl(rl relay.Relay, payload interface{}) error {
	const op = errors.Op("send_control")

	fr := getFrame()
	defer putFrame(fr)

	fr.WriteVersion(frame.VERSION_1)
	fr.WriteFlags(frame.CONTROL)

	if data, ok := payload.([]byte); ok {
		// check if payload no more that 4Gb
		if uint32(len(data)) > ^uint32(0) {
			return errors.E(op, errors.Str("payload is more that 4gb"))
		}

		fr.WritePayloadLen(uint32(len(data)))
		fr.WritePayload(data)
		fr.WriteCRC()

		err := rl.Send(fr)
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return errors.E(op, errors.Errorf("invalid payload: %s", err))
	}

	fr.WritePayloadLen(uint32(len(data)))
	fr.WritePayload(data)
	fr.WriteCRC()

	// hold a pointer to a frame
	// Do we need a copy here????
	err = rl.Send(fr)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func FetchPID(rl relay.Relay) (int64, error) {
	const op = errors.Op("fetch_pid")
	err := SendControl(rl, pidCommand{Pid: os.Getpid()})
	if err != nil {
		return 0, errors.E(op, err)
	}

	fr := getFrame()
	defer putFrame(fr)

	err = rl.Receive(fr)
	if !fr.VerifyCRC() {
		return 0, errors.E(op, errors.Str("CRC mismatch"))
	}
	if err != nil {
		return 0, errors.E(op, err)
	}
	if fr == nil {
		return 0, errors.E(op, errors.Str("nil frame received"))
	}

	flags := fr.ReadFlags()

	if flags&frame.CONTROL == 0 {
		return 0, errors.E(op, errors.Str("unexpected response, header is missing, no CONTROL flag"))
	}

	link := &pidCommand{}
	err = json.Unmarshal(fr.Payload(), link)
	if err != nil {
		return 0, errors.E(op, err)
	}

	return int64(link.Pid), nil
}
