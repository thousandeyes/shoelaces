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
	"os"
	"path/filepath"
	"testing"
)

func TestSetFlagsAppliesDefaults(t *testing.T) {
	env := defaultEnvironment()
	if _, err := env.setFlags(nil, nil); err != nil {
		t.Fatal(err)
	}

	if env.BindAddr != "localhost:8081" {
		t.Errorf("Expected default bind address, got %q", env.BindAddr)
	}
	if env.StaticDir != "web" {
		t.Errorf("Expected default static dir, got %q", env.StaticDir)
	}
	if env.EnvDir != "env_overrides" {
		t.Errorf("Expected default env dir, got %q", env.EnvDir)
	}
	if env.TemplateExtension != ".slc" {
		t.Errorf("Expected default template extension, got %q", env.TemplateExtension)
	}
	if env.MappingsFile != "mappings.yaml" {
		t.Errorf("Expected default mappings file, got %q", env.MappingsFile)
	}
}

func TestSetFlagsLoadsConfigEnvAndCLIInOrder(t *testing.T) {
	configFile := writeConfig(t, "shoelaces.conf", ""+
		"# comment\n"+
		"bind-addr=config:1\n"+
		"data-dir config-data\n"+
		"static-dir=config-static\n"+
		"debug\n")

	env := defaultEnvironment()
	args := []string{"-bind-addr", "cli:3", "-mappings-file", "cli.yaml"}
	environ := []string{
		"CONFIG=" + configFile,
		"BIND_ADDR=env:2",
		"DEBUG=false",
		"STATIC_DIR=env-static",
	}

	if _, err := env.setFlags(args, environ); err != nil {
		t.Fatal(err)
	}

	if env.ConfigFile != configFile {
		t.Errorf("Expected config file %q, got %q", configFile, env.ConfigFile)
	}
	if env.BindAddr != "cli:3" {
		t.Errorf("Expected CLI bind address, got %q", env.BindAddr)
	}
	if env.DataDir != "config-data" {
		t.Errorf("Expected data dir from config, got %q", env.DataDir)
	}
	if env.StaticDir != "env-static" {
		t.Errorf("Expected static dir from env, got %q", env.StaticDir)
	}
	if env.MappingsFile != "cli.yaml" {
		t.Errorf("Expected mappings file from CLI, got %q", env.MappingsFile)
	}
	if env.Debug {
		t.Error("Expected DEBUG env var to override bare debug config value")
	}
}

func TestSetFlagsCLIConfigOverridesEnvConfig(t *testing.T) {
	cliConfig := writeConfig(t, "cli.conf", "data-dir=cli-data\n")
	envConfig := writeConfig(t, "env.conf", "data-dir=env-data\n")

	env := defaultEnvironment()
	if _, err := env.setFlags([]string{"-config", cliConfig}, []string{"CONFIG=" + envConfig}); err != nil {
		t.Fatal(err)
	}

	if env.ConfigFile != cliConfig {
		t.Errorf("Expected CLI config file %q, got %q", cliConfig, env.ConfigFile)
	}
	if env.DataDir != "cli-data" {
		t.Errorf("Expected data dir from CLI-selected config, got %q", env.DataDir)
	}
}

func TestSetFlagsEmptyEnvOverridesConfigValue(t *testing.T) {
	configFile := writeConfig(t, "shoelaces.conf", "base-url=config-base-url\n")

	env := defaultEnvironment()
	if _, err := env.setFlags(nil, []string{"CONFIG=" + configFile, "BASE_URL="}); err != nil {
		t.Fatal(err)
	}

	if env.BaseURL != "" {
		t.Errorf("Expected empty base URL from env, got %q", env.BaseURL)
	}
}

func TestSetFlagsEmptyEnvOverridesDefaultValue(t *testing.T) {
	env := defaultEnvironment()
	if _, err := env.setFlags(nil, []string{"ENV_DIR="}); err != nil {
		t.Fatal(err)
	}

	if env.EnvDir != "" {
		t.Errorf("Expected empty env dir from env, got %q", env.EnvDir)
	}
}

func TestSetFlagsEmptyConfigValueOverridesDefault(t *testing.T) {
	configFile := writeConfig(t, "shoelaces.conf", "env-dir=\n")

	env := defaultEnvironment()
	if _, err := env.setFlags([]string{"-config", configFile}, nil); err != nil {
		t.Fatal(err)
	}

	if env.EnvDir != "" {
		t.Errorf("Expected empty env dir from config, got %q", env.EnvDir)
	}
}

func TestSetFlagsEmptyCLIConfigSkipsEnvConfig(t *testing.T) {
	envConfig := writeConfig(t, "env.conf", "base-url=env-config-base-url\n")

	env := defaultEnvironment()
	if _, err := env.setFlags([]string{"-config="}, []string{"CONFIG=" + envConfig}); err != nil {
		t.Fatal(err)
	}

	if env.BaseURL != "" {
		t.Errorf("Expected env config file to be skipped, got base URL %q", env.BaseURL)
	}
	if env.ConfigFile != "" {
		t.Errorf("Expected empty config file from CLI, got %q", env.ConfigFile)
	}
}

func TestSetFlagsEmptyBoolEnvMeansTrue(t *testing.T) {
	configFile := writeConfig(t, "shoelaces.conf", "debug=false\n")

	env := defaultEnvironment()
	if _, err := env.setFlags(nil, []string{"CONFIG=" + configFile, "DEBUG="}); err != nil {
		t.Fatal(err)
	}

	if !env.Debug {
		t.Error("Expected empty DEBUG env var to set debug true")
	}
}

func TestSetFlagsReturnsConfigErrors(t *testing.T) {
	tests := []struct {
		name   string
		config string
	}{
		{name: "unknown key", config: "unknown=value\n"},
		{name: "invalid bool", config: "debug=maybe\n"},
		{name: "empty bool", config: "debug=\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := defaultEnvironment()
			configFile := writeConfig(t, "shoelaces.conf", tt.config)

			if _, err := env.setFlags([]string{"-config", configFile}, nil); err == nil {
				t.Fatal("Expected error")
			}
		})
	}
}

func TestValidateFlagsReturnsError(t *testing.T) {
	env := defaultEnvironment()
	env.DataDir = "data"
	env.StaticDir = "web"
	if err := env.validateFlags(); err != nil {
		t.Fatal(err)
	}

	env.StaticDir = ""
	if err := env.validateFlags(); err == nil {
		t.Fatal("Expected error")
	}
}

func writeConfig(t *testing.T, name string, contents string) string {
	t.Helper()

	configFile := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(configFile, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return configFile
}
