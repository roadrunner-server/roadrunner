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
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
)

var stopSignal = make(chan os.Signal, 1)
var source string

func init() {
	CLI.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Serve RoadRunner service(s)",
		RunE:  serveHandler,
	})
	CLI.PersistentFlags().StringVarP(&source, "source", "s", "", "Source directory")

	signal.Notify(stopSignal, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
}

func serveHandler(cmd *cobra.Command, args []string) error {
	if Debug && source != "" {
		log.Printf("Watch files")

		w := watcher.New()
		w.FilterOps(watcher.Rename, watcher.Create, watcher.Move, watcher.Write)

		go func() {
			for {
				select {
				case event := <-w.Event:
					log.Println(event)
					serveHandler(cmd, args)
				case err := <-w.Error:
					log.Fatalln(err)
				case <-w.Closed:
					return
				}
			}
		}()

		if err := w.AddRecursive(source); err != nil {
			log.Fatalln(err)
		}

		for path, f := range w.WatchedFiles() {
			fmt.Printf("%s: %s\n", path, f.Name())
		}

		go w.Start(100 * time.Millisecond)
	}

	stopped := make(chan interface{})

	go func() {
		<-stopSignal
		Container.Stop()
		close(stopped)
	}()

	if err := Container.Serve(); err != nil {
		return err
	}

	<-stopped
	return nil
}
