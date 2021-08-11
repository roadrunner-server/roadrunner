package jobs

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	pq "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type Type uint32

const (
	Error Type = iota
	NoError
)

// internal worker protocol (jobs mode)
type protocol struct {
	// message type, see Type
	T Type `json:"type"`
	// Payload
	Data []byte `json:"data"`
}

type errorResp struct {
	Msg     string              `json:"message"`
	Requeue bool                `json:"requeue"`
	Delay   int64               `json:"delay_seconds"`
	Headers map[string][]string `json:"headers"`
}

func handleResponse(resp []byte, jb pq.Item, log logger.Logger) error {
	const op = errors.Op("jobs_handle_response")
	// TODO(rustatian) to sync.Pool
	p := &protocol{}

	err := json.Unmarshal(resp, p)
	if err != nil {
		return errors.E(op, err)
	}

	switch p.T {
	// likely case
	case NoError:
		err = jb.Ack()
		if err != nil {
			return errors.E(op, err)
		}
	case Error:
		// TODO(rustatian) to sync.Pool
		er := &errorResp{}

		err = json.Unmarshal(p.Data, er)
		if err != nil {
			return errors.E(op, err)
		}

		log.Error("error protocol type", "error", er.Msg, "delay", er.Delay, "requeue", er.Requeue)

		if er.Requeue {
			err = jb.Requeue(er.Headers, er.Delay)
			if err != nil {
				return errors.E(op, err)
			}
			return nil
		}
	default:
		err = jb.Ack()
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}
