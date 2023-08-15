// Copyright 2014 Docker, Inc.
// Copyright 2015-2018 CoreOS, Inc.
// Copyright 2023 SpiralScout.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package sdnotify provides a Go implementation of the sd_notify protocol.
// It can be used to inform systemd of service start-up completion, watchdog
// events, and other status changes.
//
// https://www.freedesktop.org/software/systemd/man/sd_notify.html#Description
package sdnotify

import (
	"net"
	"os"
	"time"
)

type State string

const (
	// Ready tells the service manager that service startup is finished
	// or the service finished loading its configuration.
	// https://www.freedesktop.org/software/systemd/man/sd_notify.html#READY=1
	Ready State = "READY=1"

	// Stopping tells the service manager that the service is beginning
	// its shutdown.
	Stopping State = "STOPPING=1"

	// Reloading tells the service manager that this service is
	// reloading its configuration. Note that you must call SdNotifyReady when
	// it completed reloading.
	Reloading State = "RELOADING=1"

	// Watchdog tells the service manager to update the watchdog
	// timestamp for the service.
	Watchdog = "WATCHDOG=1"
)

// SdNotify sends a message to the init daemon. It is common to ignore the error.
// If `unsetEnvironment` is true, the environment variable `NOTIFY_SOCKET`
// will be unconditionally unset.
//
// It returns one of the following:
// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET or while sending data)
// (true, nil) - notification supported, data has been sent
func SdNotify(state State) (bool, error) {
	socketAddr := &net.UnixAddr{
		Name: os.Getenv("NOTIFY_SOCKET"),
		Net:  "unixgram",
	}

	// NOTIFY_SOCKET not set
	if socketAddr.Name == "" {
		return false, nil
	}

	conn, err := net.DialUnix(socketAddr.Net, nil, socketAddr)
	// Error connecting to NOTIFY_SOCKET
	if err != nil {
		return false, err
	}

	defer func() {
		_ = conn.Close()
	}()

	if _, err = conn.Write([]byte(state)); err != nil {
		return false, err
	}

	return true, nil
}

func StartWatchdog(interval int, stopCh chan struct{}) {
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				supported, err := SdNotify(Watchdog)
				if err != nil {
					// notification supported, but failure happened, retry
					continue
				}
				// notification not supported, stop
				if !supported && err == nil {
					return
				}
				// notification supported, data has been sent, continue
				if supported && err == nil {
					continue
				}
			}
		}
	}()
}
