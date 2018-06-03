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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/cmd/rr/utils"
	"os"
	"fmt"
)

// Service bus for all the commands.
var (
	// Shared service bus.
	Bus *service.Bus

	// Root is application endpoint.
	Root = &cobra.Command{
		Use:   "rr",
		Short: "RoadRunner, PHP application server",
	}

	cfgFile string
	verbose bool
)

// Execute adds all child commands to the Root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the Root.
func Execute(serviceBus *service.Bus) {
	Bus = serviceBus
	if err := Root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	Root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	Root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .rr.yaml)")

	cobra.OnInitialize(func() {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if cfg := initConfig(cfgFile, []string{"."}, ".rr"); cfg != nil {
			if err := Bus.Configure(cfg); err != nil {
				panic(err)
			}
		}
	})
}

func initConfig(cfgFile string, path []string, name string) service.Config {
	cfg := viper.New()

	if cfgFile != "" {
		// Use cfg file from the flag.
		cfg.SetConfigFile(cfgFile)
	} else {
		// automatic location
		for _, p := range path {
			cfg.AddConfigPath(p)
		}
		cfg.SetConfigName(name)
	}

	// read in environment variables that match
	cfg.AutomaticEnv()

	// If a cfg file is found, read it in.
	if err := cfg.ReadInConfig(); err != nil {
		logrus.Warnf("config: %s", err)
		return nil
	}

	return &utils.ConfigWrapper{cfg}
}
