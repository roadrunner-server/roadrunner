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
	"log"
	"time"
	"github.com/sirupsen/logrus"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use: "serve",
		Run: serveHandler,
	})
}

func serveHandler(cmd *cobra.Command, args []string) {
	r := roadrunner.NewRouter(
		func() *exec.Cmd {
			return exec.Command("php", "/Users/wolfy-j/Projects/phpapp/webroot/index.php", "rr", "pipes")
		},
		roadrunner.NewPipeFactory(),
	)

	err := r.Configure(roadrunner.Config{
		NumWorkers:      1,
		AllocateTimeout: time.Minute,
		DestroyTimeout:  time.Minute,
	})

	r.Observe(func(event int, ctx interface{}) {
		logrus.Info(ctx)
	})

	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		r.Exec(&roadrunner.Payload{})
	}

	log.Print(r.Workers())

}
