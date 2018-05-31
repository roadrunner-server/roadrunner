// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spiral/roadrunner"
	"os/exec"
	"time"
	"github.com/sirupsen/logrus"
	rrhttp "github.com/spiral/roadrunner/psr7"
	"net/http"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use: "serve",
		Run: serveHandler,
	})
}

func serveHandler(cmd *cobra.Command, args []string) {
	rr := roadrunner.NewRouter(
		func() *exec.Cmd {
			return exec.Command("php", "/Users/wolfy-j/Projects/phpapp/webroot/index.php", "rr", "pipes")
		},
		roadrunner.NewPipeFactory(),
	)

	err := rr.Configure(roadrunner.Config{
		NumWorkers:      4,
		AllocateTimeout: time.Minute,
		DestroyTimeout:  time.Minute,
		//MaxExecutions:   10,
	})

	rr.Observe(func(event int, ctx interface{}) {
		logrus.Info(ctx)
	})

	if err != nil {
		panic(err)
	}

	logrus.Info("serving")

	//Enable http2
	//srv := http.Server{
	//	Addr: ":8080",
	//	Handler: rrhttp.NewServer(
	//		rrhttp.Config{
	//			ServeStatic: true,
	//			Root:        "/Users/wolfy-j/Projects/phpapp/webroot",
	//		},
	//		rr,
	//	),
	//}

	//	srv.ListenAndServe()

	//http2.ConfigureServer(&srv, nil)
	//srv.ListenAndServeTLS("localhost.cert", "localhost.key")

	http.ListenAndServe(":8080", rrhttp.NewServer(
		rrhttp.Config{
			ServeStatic: true,
			Root:        "/Users/wolfy-j/Projects/phpapp/webroot",
		},
		rr,
	))
}
