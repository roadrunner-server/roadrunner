// Copyright (c) 2018 SpiralScout
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package http

import (
	"github.com/spf13/cobra"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	"github.com/go-errors/errors"
	"github.com/spiral/roadrunner/rpc"
)

func init() {
	rr.CLI.AddCommand(&cobra.Command{
		Use:   "http:reload",
		Short: "Reload RoadRunner worker pools for the HTTP service",
		RunE:  reloadHandler,
	})
}

func reloadHandler(cmd *cobra.Command, args []string) error {
	if !rr.Services.Has("rpc") {
		return errors.New("RPC service is not configured")
	}

	client, err := rr.Services.Get("rpc").(*rpc.Service).Client()
	if err != nil {
		return err
	}
	defer client.Close()

	var r string
	if err := client.Call("http.Reset", true, &r); err != nil {
		return err
	}

	rr.Logger.Info("http.service: restarting worker pool")
	return nil
}
