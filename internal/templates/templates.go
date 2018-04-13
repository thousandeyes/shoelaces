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

package templates

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/thousandeyes/shoelaces/internal/log"
	"github.com/thousandeyes/shoelaces/internal/utils"
)

const defaultEnvironment = "default"

var varRegex = regexp.MustCompile(`{{\.(.*?)}}`)
var configNameRegex = regexp.MustCompile(`{{define\s+"(.*?)".*}}`)

// ShoelacesTemplates holds the core attributes for handling the dyanmic configurations
// in Shoelaces.
type ShoelacesTemplates struct {
	envTemplates map[string]shoelacesTemplateEnvironment
	dataDir      string
	envDir       string
	tplExt       string
}

type shoelacesTemplateEnvironment struct {
	templateObj  *template.Template
	templateVars map[string][]string
}

type shoelacesTemplateInfo struct {
	name      string
	variables []string
}

// New creates and initializes a new ShoelacesTemplates instance a returns a pointer to
// it.
func New() *ShoelacesTemplates {
	e := make(map[string]shoelacesTemplateEnvironment)
	e[defaultEnvironment] = shoelacesTemplateEnvironment{
		templateObj:  template.New(""),
		templateVars: make(map[string][]string),
	}
	return &ShoelacesTemplates{envTemplates: e}
}

func (s *ShoelacesTemplates) parseTemplateInfo(logger log.Logger, path string) shoelacesTemplateInfo {
	fh, err := os.Open(path)
	if err != nil {
		logger.Error("component", "template", "err", err.Error())
		os.Exit(1)
	}

	defer fh.Close()

	templateVars := make([]string, 0)
	scanner := bufio.NewScanner(fh)
	templateName := ""
	i := 0
	for scanner.Scan() {
		// find variables
		result := varRegex.FindAllStringSubmatch(scanner.Text(), -1)
		if varRegex.MatchString(scanner.Text()) {
			for _, v := range result {
				// we only want the actual match, being second in the group
				if !utils.StringInSlice(v[1], templateVars) {
					templateVars = append(templateVars, v[1])
				}
			}
		}
		// if first line get name of template
		if i == 0 {
			nameResult := configNameRegex.FindAllStringSubmatch(scanner.Text(), -1)
			templateName = nameResult[0][1]
		}
		i++
	}

	return shoelacesTemplateInfo{name: templateName, variables: templateVars}
}

func (s *ShoelacesTemplates) checkAddEnvironment(logger log.Logger, environment string) {
	if _, ok := s.envTemplates[environment]; !ok {
		c, e := s.envTemplates[defaultEnvironment].templateObj.Clone()
		if e != nil {
			logger.Error("component", "template", "msg", "Template for environment already executed", "environment", environment)
			os.Exit(1)
		}
		s.envTemplates[environment] = shoelacesTemplateEnvironment{
			templateObj:  c,
			templateVars: make(map[string][]string),
		}
	}
}

func (s *ShoelacesTemplates) addTemplate(logger log.Logger, path string, environment string) error {
	s.checkAddEnvironment(logger, environment)
	i := s.parseTemplateInfo(logger, path)
	_, err := s.envTemplates[environment].templateObj.ParseFiles(path)
	if err != nil {
		return err
	}
	s.envTemplates[environment].templateVars[i.name] = i.variables
	return nil
}

func (s *ShoelacesTemplates) getEnvFromPath(path string) string {
	envPath := filepath.Join(s.dataDir, s.envDir)
	if strings.HasPrefix(path, envPath) {
		return strings.Split(strings.TrimPrefix(path, envPath), "/")[1]
	}
	return defaultEnvironment
}

// ParseTemplates travels the dataDir and loads in an internal structure
// all the templates found.
func (s *ShoelacesTemplates) ParseTemplates(logger log.Logger, dataDir string, envDir string, envs []string, tplExt string) {
	s.dataDir = dataDir
	s.envDir = envDir
	s.tplExt = tplExt

	logger.Debug("component", "template", "msg", "Template parsing started", "dir", dataDir)

	tplScannerDefault := func(p string, info os.FileInfo, err error) error {
		if strings.HasPrefix(p, path.Join(dataDir, envDir)) {
			return err
		}
		if strings.HasSuffix(p, tplExt) {
			logger.Info("component", "template", "msg", "Parsing file", "file", p)
			if err := s.addTemplate(logger, p, defaultEnvironment); err != nil {
				logger.Error("component", "template", "err", err.Error())
				os.Exit(1)
			}
		}
		return err
	}

	tplScannerOverride := func(p string, info os.FileInfo, err error) error {
		if strings.HasSuffix(p, tplExt) {
			env := s.getEnvFromPath(p)
			logger.Info("component", "template", "msg", "Parsing ovveride", "environment", env, "file", p)

			if err := s.addTemplate(logger, p, env); err != nil {
				logger.Error("component", "template", "err", err.Error())
				os.Exit(1)
			}
		}
		return err
	}

	if err := filepath.Walk(dataDir, tplScannerDefault); err != nil {
		panic(err)
	}
	logger.Info("component", "template", "msg", "Parsing override files", "dir", path.Join(dataDir, envDir))
	if err := filepath.Walk(path.Join(dataDir, envDir), tplScannerOverride); err != nil {
		logger.Info("component", "template", "msg", "No overrides found")
	}
	logger.Debug("component", "template", "msg", "Parsing ended")
}

// RenderTemplate receives a name and a map of parameters, among other
// arguments, and returns the rendered template. It's aware of the
// environment, in case of any.
func (s *ShoelacesTemplates) RenderTemplate(logger log.Logger, configName string, paramMap map[string]interface{}, envName string) (string, error) {
	if envName == "" {
		envName = defaultEnvironment
	}
	logger.Info("component", "template", "action", "template-request", "template", configName, "env", envName, "parameters", utils.MapToString(paramMap))

	requiredVariables := s.envTemplates[envName].templateVars[configName]

	var b bytes.Buffer
	err := s.envTemplates[envName].templateObj.ExecuteTemplate(&b, configName, paramMap)
	// Fall back to default template in case this is non default environment
	// XXX: this is temporary and will be simplified to reduce the code duplication
	if err != nil && envName != defaultEnvironment {
		requiredVariables = s.envTemplates[defaultEnvironment].templateVars[configName]
		err = s.envTemplates[defaultEnvironment].templateObj.ExecuteTemplate(&b, configName, paramMap)
	}
	if err != nil {
		logger.Info("component", "template", "action", "render-template", "err", err.Error())
		return "", err
	}
	r := b.String()
	if strings.Contains(r, "<no value>") {
		missingVariables := ""
		for _, requiredVariable := range requiredVariables {
			if !utils.KeyInMap(requiredVariable, paramMap) {
				if len(missingVariables) > 0 {
					missingVariables += ", "
				}
				missingVariables += requiredVariable
			}
		}
		logger.Info("component", "template", "msg", "Missing variables in request", "variables", missingVariables)
		return "", errors.New("Missing variables in request: " + missingVariables)
	}

	return r, nil
}

// ListVariables receives a template name and return the list of variables
// that belong to it. It's mainly used by the web frontend to provide a
// list of dynamic fields to complete before rendering a template.
func (s *ShoelacesTemplates) ListVariables(templateName, envName string) []string {
	if e, ok := s.envTemplates[envName]; ok {
		if v, ok := e.templateVars[templateName]; ok {
			return v
		}
	}
	var empty []string
	return empty
}
