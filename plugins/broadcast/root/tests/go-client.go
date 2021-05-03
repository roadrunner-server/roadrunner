package main

import (
	"fmt"
	"os"

	"github.com/spiral/broadcast/v2"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	"github.com/spiral/roadrunner/service/rpc"
	"golang.org/x/sync/errgroup"
)

type logService struct {
	broadcast *broadcast.Service
	stop      chan interface{}
}

func (l *logService) Init(service *broadcast.Service) (bool, error) {
	l.broadcast = service

	return true, nil
}

func (l *logService) Serve() error {
	l.stop = make(chan interface{})

	client := l.broadcast.NewClient()
	if err := client.SubscribePattern("tests/*"); err != nil {
		return err
	}

	logFile, _ := os.Create("log.txt")

	g := &errgroup.Group{}
	g.Go(func() error {
		for msg := range client.Channel() {
			_, err := logFile.Write([]byte(fmt.Sprintf(
				"%s: %s\n",
				msg.Topic,
				string(msg.Payload),
			)))
			if err != nil {
				return err
			}

			err = logFile.Sync()
			if err != nil {
				return err
			}
		}
		return nil
	})

	<-l.stop
	err := logFile.Close()
	if err != nil {
		return err
	}

	err = client.Close()
	if err != nil {
		return err
	}

	return g.Wait()
}

func (l *logService) Stop() {
	close(l.stop)
}

func main() {
	rr.Container.Register(rpc.ID, &rpc.Service{})
	rr.Container.Register(broadcast.ID, &broadcast.Service{})
	rr.Container.Register("log", &logService{})

	rr.Execute()
}
