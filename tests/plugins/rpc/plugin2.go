package rpc

import (
	"net"
	"net/rpc"
	"time"

	"github.com/spiral/errors"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
)

// plugin2 makes a call to the plugin1 via RPC
// this is just a simulation of external call FOR TEST
// you don't need to do such things :)
type Plugin2 struct {
}

func (p2 *Plugin2) Init() error {
	return nil
}

func (p2 *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)

	go func() {
		time.Sleep(time.Second * 3)

		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			errCh <- errors.E(errors.Serve, err)
			return
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		var ret string
		err = client.Call("rpc_test.plugin1.Hello", "Valery", &ret)
		if err != nil {
			errCh <- err
			return
		}
		if ret != "Hello, username: Valery" {
			errCh <- errors.E("wrong response")
			return
		}
		// to stop exec
		errCh <- errors.E(errors.Disabled)
	}()

	return errCh
}

func (p2 *Plugin2) Stop() error {
	return nil
}
