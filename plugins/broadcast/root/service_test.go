package broadcast

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
)

var rpcPort = 6010

func setup(cfg string) (*Service, *rpc.Service, service.Container) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	err := c.Init(&testCfg{
		broadcast: cfg,
		rpc:       fmt.Sprintf(`{"listen":"tcp://:%v"}`, rpcPort),
	})

	rpcPort++

	if err != nil {
		panic(err)
	}

	go func() {
		err = c.Serve()
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Millisecond * 100)

	b, _ := c.Get(ID)
	br := b.(*Service)

	r, _ := c.Get(rpc.ID)
	rp := r.(*rpc.Service)

	return br, rp, c
}

func readStr(m *Message) string {
	return strings.TrimRight(string(m.Payload), "\n")
}

func newMessage(t, m string) *Message {
	return &Message{Topic: t, Payload: []byte(m)}
}

func TestService_Publish(t *testing.T) {
	svc := &Service{}
	assert.Error(t, svc.Publish(nil))
}
