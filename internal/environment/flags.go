// Copyright 2018-2026 ThousandEyes Inc.
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
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (env *Environment) setFlags(args []string, environ []string) (*flag.FlagSet, error) {
	env.setFlagDefaults()

	configFile, configFromArgs := configFileFromArgs(args)
	if !configFromArgs {
		configFile, _ = envValue(environ, "CONFIG")
	}
	if configFile != "" {
		env.ConfigFile = configFile
		if err := env.loadConfigFile(configFile); err != nil {
			return nil, err
		}
	}

	if err := env.applyEnvVars(environ); err != nil {
		return nil, err
	}

	flags := env.registerFlags()
	if err := flags.Parse(args); err != nil {
		return flags, err
	}
	return flags, nil
}

func (env *Environment) setFlagDefaults() {
	env.BindAddr = "localhost:8081"
	env.StaticDir = "web"
	env.EnvDir = "env_overrides"
	env.TemplateExtension = ".slc"
	env.MappingsFile = "mappings.yaml"
}

func (env *Environment) registerFlags() *flag.FlagSet {
	flags := flag.NewFlagSet("shoelaces", flag.ContinueOnError)
	flags.StringVar(&env.ConfigFile, "config", env.ConfigFile, "My config file")
	flags.StringVar(&env.BindAddr, "bind-addr", env.BindAddr, "The address where I'm going to listen")
	flags.StringVar(&env.BaseURL, "base-url", env.BaseURL, "The base shoelaces URL. If it's not defined, it will default to bind-addr.")
	flags.StringVar(&env.DataDir, "data-dir", env.DataDir, "Directory with mappings, configs, templates, etc.")
	flags.StringVar(&env.StaticDir, "static-dir", env.StaticDir, "A custom web directory with static files")
	flags.StringVar(&env.EnvDir, "env-dir", env.EnvDir, "Directory with overrides")
	flags.StringVar(&env.TemplateExtension, "template-extension", env.TemplateExtension, "Shoelaces template extension")
	flags.StringVar(&env.MappingsFile, "mappings-file", env.MappingsFile, "My mappings YAML file")
	flags.BoolVar(&env.Debug, "debug", env.Debug, "Debug mode")
	return flags
}

func configFileFromArgs(args []string) (string, bool) {
	for i, arg := range args {
		if arg == "-config" || arg == "--config" {
			if i+1 < len(args) {
				return args[i+1], true
			}
			return "", true
		}
		if strings.HasPrefix(arg, "-config=") {
			return strings.TrimPrefix(arg, "-config="), true
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config="), true
		}
	}
	return "", false
}

func (env *Environment) loadConfigFile(configFile string) error {
	contents, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(contents), "\n") {
		key, value, ok := parseConfigLine(line)
		if !ok {
			continue
		}
		if err := env.setConfigValue(key, value); err != nil {
			return err
		}
	}
	return nil
}

func parseConfigLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}

	if key, value, found := strings.Cut(line, "="); found {
		return strings.TrimSpace(key), strings.TrimSpace(value), true
	}

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", "", false
	}
	if len(fields) == 1 {
		return fields[0], "true", true
	}
	return fields[0], strings.Join(fields[1:], " "), true
}

func (env *Environment) applyEnvVars(environ []string) error {
	if err := env.applyEnvVar(environ, "config", "CONFIG"); err != nil {
		return err
	}
	if err := env.applyEnvVar(environ, "bind-addr", "BIND_ADDR"); err != nil {
		return err
	}
	if err := env.applyEnvVar(environ, "base-url", "BASE_URL"); err != nil {
		return err
	}
	if err := env.applyEnvVar(environ, "data-dir", "DATA_DIR"); err != nil {
		return err
	}
	if err := env.applyEnvVar(environ, "static-dir", "STATIC_DIR"); err != nil {
		return err
	}
	if err := env.applyEnvVar(environ, "env-dir", "ENV_DIR"); err != nil {
		return err
	}
	if err := env.applyEnvVar(environ, "template-extension", "TEMPLATE_EXTENSION"); err != nil {
		return err
	}
	if err := env.applyEnvVar(environ, "mappings-file", "MAPPINGS_FILE"); err != nil {
		return err
	}
	return env.applyEnvVar(environ, "debug", "DEBUG")
}

func (env *Environment) applyEnvVar(environ []string, key, name string) error {
	value, ok := envValue(environ, name)
	if !ok {
		return nil
	}
	if key == "debug" && value == "" {
		value = "true"
	}
	return env.setConfigValue(key, value)
}

func envValue(environ []string, name string) (string, bool) {
	for _, entry := range environ {
		key, value, found := strings.Cut(entry, "=")
		if found && key == name {
			return value, true
		}
	}
	return "", false
}

func (env *Environment) setConfigValue(key, value string) error {
	switch key {
	case "config":
		env.ConfigFile = value
	case "bind-addr":
		env.BindAddr = value
	case "base-url":
		env.BaseURL = value
	case "data-dir":
		env.DataDir = value
	case "static-dir":
		env.StaticDir = value
	case "env-dir":
		env.EnvDir = value
	case "template-extension":
		env.TemplateExtension = value
	case "mappings-file":
		env.MappingsFile = value
	case "debug":
		debug, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid debug value %q: %w", value, err)
		}
		env.Debug = debug
	default:
		return fmt.Errorf("unknown config key %q", key)
	}
	return nil
}

func (env *Environment) validateFlags() error {
	var messages []string

	if env.DataDir == "" {
		messages = append(messages, "[*] You must specify the data-dir parameter")
	}

	if env.StaticDir == "" {
		messages = append(messages, "[*] You must specify the data-dir parameter")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, "\n"))
	}
	return nil
}
