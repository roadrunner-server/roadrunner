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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spiral/roadrunner/cmd/util"
)

func init() {
	CLI.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop RoadRunner server",
		RunE:  stopHandler,
	})
}

func stopHandler(cmd *cobra.Command, args []string) error {
	client, err := util.RPCClient(Container)
	if err != nil {
		return err
	}
	defer client.Close()

	util.Printf("<green>Stopping RoadRunner</reset>: ")

	var r string
	if err := client.Call("system.Stop", true, &r); err != nil {
		return err
	}

	util.Printf("<green+hb>done</reset>\n")
	return nil
}
