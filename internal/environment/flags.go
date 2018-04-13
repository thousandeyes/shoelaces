// Copyright 2018 ThousandEyes Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package environment

import (
	"fmt"
	"os"

	"github.com/namsral/flag"
)

func (env *Environment) setFlags() {
	flag.StringVar(&env.ConfigFile, "config", "", "My config file")
	flag.IntVar(&env.Port, "port", 8080, "The port where I'm going to listen")
	flag.StringVar(&env.Domain, "domain", "localhost", "The address where I'm going to listen")
	flag.StringVar(&env.DataDir, "data-dir", "", "Directory with mappings, configs, templates, etc.")
	flag.StringVar(&env.StaticDir, "static-dir", "web", "A custom web directory with static files")
	flag.StringVar(&env.EnvDir, "env-dir", "env_overrides", "Directory with overrides")
	flag.StringVar(&env.TemplateExtension, "template-extension", ".slc", "Shoelaces template extension")
	flag.StringVar(&env.MappingsFile, "mappings-file", "mappings.yaml", "My mappings YAML file")
	flag.BoolVar(&env.Debug, "debug", false, "Debug mode")

	flag.Parse()
}

func (env *Environment) validateFlags() {
	error := false

	if env.DataDir == "" {
		fmt.Println("[*] You must specify the data-dir parameter")
		error = true
	}

	if env.StaticDir == "" {
		fmt.Println("[*] You must specify the data-dir parameter")
		error = true
	}

	if error {
		fmt.Println("\nAvailable parameters:")
		flag.PrintDefaults()
		fmt.Println("\nParameters can be specified as environment variables, arguments or in a config file.")
		os.Exit(1)
	}
}
