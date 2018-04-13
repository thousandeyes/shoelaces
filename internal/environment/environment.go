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
	"html/template"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/thousandeyes/shoelaces/internal/event"
	"github.com/thousandeyes/shoelaces/internal/log"
	"github.com/thousandeyes/shoelaces/internal/mappings"
	"github.com/thousandeyes/shoelaces/internal/server"
	"github.com/thousandeyes/shoelaces/internal/templates"
)

// Environment struct holds the shoelaces instance global data.
type Environment struct {
	ConfigFile      string
	BaseURL         string
	HostnameMaps    []mappings.HostnameMap
	NetworkMaps     []mappings.NetworkMap
	ServerStates    *server.States
	EventLog        *event.Log
	ParamsBlacklist []string
	Templates       *templates.ShoelacesTemplates // Dynamic slc templates
	StaticTemplates *template.Template            // Static Templates
	Environments    []string                      // Valid config environments
	Logger          log.Logger

	Port              int
	Domain            string
	DataDir           string
	StaticDir         string
	EnvDir            string
	TemplateExtension string
	MappingsFile      string
	Debug             bool
}

// New returns an initialized environment structure
func New() *Environment {
	env := defaultEnvironment()
	env.setFlags()
	env.validateFlags()

	if env.Debug {
		env.Logger = log.AllowDebug(env.Logger)
	}

	env.BaseURL = fmt.Sprintf("%s:%d", env.Domain, env.Port)
	env.Environments = env.initEnvOverrides()

	env.EventLog = &event.Log{}

	env.Logger.Info("component", "environment", "msg", "Override found", "environment", env.Environments)

	mappingsPath := path.Join(env.DataDir, env.MappingsFile)
	if err := env.initMappings(mappingsPath); err != nil {
		panic(err)
	}

	env.initStaticTemplates()
	env.Templates.ParseTemplates(env.Logger, env.DataDir, env.EnvDir, env.Environments, env.TemplateExtension)
	server.StartStateCleaner(env.Logger, env.ServerStates)

	return env
}

func defaultEnvironment() *Environment {
	env := &Environment{}
	env.NetworkMaps = make([]mappings.NetworkMap, 0)
	env.HostnameMaps = make([]mappings.HostnameMap, 0)
	env.ServerStates = &server.States{sync.RWMutex{}, make(map[string]*server.State)}
	env.ParamsBlacklist = []string{"baseURL"}
	env.Templates = templates.New()
	env.Environments = make([]string, 0)
	env.Logger = log.MakeLogger(os.Stdout)

	return env
}

func (env *Environment) initStaticTemplates() {
	staticTemplates := []string{
		path.Join(env.StaticDir, "templates/html/header.html"),
		path.Join(env.StaticDir, "templates/html/index.html"),
		path.Join(env.StaticDir, "templates/html/events.html"),
		path.Join(env.StaticDir, "templates/html/mappings.html"),
		path.Join(env.StaticDir, "templates/html/footer.html"),
	}

	fmt.Println(env.StaticDir)

	for _, t := range staticTemplates {
		if _, err := os.Stat(t); err != nil {
			env.Logger.Error("component", "environment", "msg", "Template does not exists!", "environment", t)
			os.Exit(1)
		}
	}

	env.StaticTemplates = template.Must(template.ParseFiles(staticTemplates...))
}

func (env *Environment) initEnvOverrides() []string {
	var environments = make([]string, 0)
	envPath := filepath.Join(env.DataDir, env.EnvDir)
	files, err := ioutil.ReadDir(envPath)
	if err == nil {
		for _, f := range files {
			if f.IsDir() {
				environments = append(environments, f.Name())
			}
		}
	}
	return environments
}

func (env *Environment) initMappings(mappingsPath string) error {
	configMappings := mappings.ParseYamlMappings(env.Logger, mappingsPath)

	for _, configNetMap := range configMappings.NetworkMaps {
		_, ipnet, err := net.ParseCIDR(configNetMap.Network)
		if err != nil {
			return err
		}

		netMap := mappings.NetworkMap{Network: ipnet, Script: initScript(configNetMap.Script)}
		env.NetworkMaps = append(env.NetworkMaps, netMap)
	}

	for _, configHostMap := range configMappings.HostnameMaps {
		regex, err := regexp.Compile(configHostMap.Hostname)
		if err != nil {
			return err
		}

		hostMap := mappings.HostnameMap{Hostname: regex, Script: initScript(configHostMap.Script)}
		env.HostnameMaps = append(env.HostnameMaps, hostMap)
	}

	return nil
}

func initScript(configScript mappings.YamlScript) *mappings.Script {
	mappingScript := &mappings.Script{
		Name:        configScript.Name,
		Environment: configScript.Environment,
		Params:      make(map[string]interface{}),
	}
	for key := range configScript.Params {
		mappingScript.Params[key] = configScript.Params[key]
	}

	return mappingScript
}
