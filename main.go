// Copyright 2018 ThousandEyes Inc.
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

package main

import (
	"net/http"
	"os"

	"github.com/thousandeyes/shoelaces/internal/environment"
	"github.com/thousandeyes/shoelaces/internal/handlers"
	"github.com/thousandeyes/shoelaces/internal/router"

	// embedded TFTP server module
	"github.com/thousandeyes/shoelaces/internal/tftpserver"
)

func main() {
	env := environment.New()
	app := handlers.MiddlewareChain(env).Then(router.ShoelacesRouter(env))

	// start embedded TFTP server if enabled (values come from environment package:
	// config file, env vars, or CLI flags defined there).
	if env.TFTP != nil && env.TFTP.Enabled {
		tf := tftpserver.New(env.TFTP.Addr, env.TFTP.Root, env.TFTP.Readonly, env.TFTP.Timeout)
		go func() {
			env.Logger.Info(
				"component", "tftp",
				"transport", "udp",
				"addr", env.TFTP.Addr,
				"root", env.TFTP.Root,
				"readonly", env.TFTP.Readonly,
				"msg", "listening",
			)
			if err := tf.ListenAndServe(); err != nil {
				env.Logger.Error("component", "tftp", "err", err)
			}
		}()
	}

	env.Logger.Info("component", "main", "transport", "http", "addr", env.BindAddr, "msg", "listening")
	env.Logger.Error("component", "main", "err", http.ListenAndServe(env.BindAddr, app))
	os.Exit(1)
}
