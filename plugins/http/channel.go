package http

import (
	"net/http"
)

// messages method used to read messages from the ws plugin with the auth requests for the topics and server
func (p *Plugin) messages() {
	for msg := range p.hub.ReceiveCh() {
		p.RLock()
		// msg here is the structure with http.ResponseWriter and http.Request
		rmsg := msg.(struct {
			RW  http.ResponseWriter
			Req *http.Request
		})

		p.handler.ServeHTTP(rmsg.RW, rmsg.Req)

		p.hub.SendCh() <- struct {
			RW  http.ResponseWriter
			Req *http.Request
		}{
			rmsg.RW,
			rmsg.Req,
		}
		p.RUnlock()
	}
}
