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

package ipxe

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/thousandeyes/shoelaces/internal/environment"
	"github.com/thousandeyes/shoelaces/internal/log"
)

// ScriptName keeps the name of a script
type ScriptName string

// EnvName holds the name of an environment
type EnvName string

// ScriptPath holds the path of a script
type ScriptPath string

// Script holds information regarding an IPXE script.
type Script struct {
	Name ScriptName
	Env  EnvName
	Path ScriptPath
}

// ScriptList receives the global environment and return a list of IPXE
// scripts.
func ScriptList(env *environment.Environment) []Script {
	ipxeScripts := make([]Script, 0)
	// Collect scripts from the main config dir.
	ipxeScripts = appendScriptsFromDir(env.Logger, ipxeScripts, env.TemplateExtension,
		filepath.Join(env.DataDir, "ipxe"), "", "/configs/")

	// Collect scripts from the config environments if any
	if len(env.Environments) > 0 {
		for _, e := range env.Environments {
			ep := filepath.Join(env.DataDir, env.EnvDir, e, "ipxe")
			ipxeScripts = appendScriptsFromDir(env.Logger, ipxeScripts, env.TemplateExtension, ep,
				EnvName(e), ScriptPath("/env/"+e+"/configs/"))
		}
	}
	return ipxeScripts
}

func appendScriptsFromDir(logger log.Logger, scripts []Script, templateExtension string, dir string, e EnvName, p ScriptPath) []Script {
	for _, s := range scriptDirList(logger, templateExtension, dir) {
		scripts = append(scripts, Script{Name: s, Env: e, Path: p})
	}
	return scripts
}

// scriptDirList returns the names of all available ipxe script templates
func scriptDirList(logger log.Logger, templateExtension string, datadir string) []ScriptName {
	files, err := ioutil.ReadDir(datadir)
	if err != nil {
		logger.Info("component=ipxescript action=dir-list dir=%s err=\"%v\"", datadir, err.Error())
		return nil
	}

	ipxeSuffix := ".ipxe"
	suffix := ipxeSuffix + templateExtension
	var pxeFiles []ScriptName
	for _, f := range files {
		// Skip over directories and non-template files.
		if f.IsDir() || !strings.HasSuffix(f.Name(), suffix) {
			continue
		}
		pxeFiles = append(pxeFiles, ScriptName(strings.TrimSuffix(f.Name(), templateExtension)))
	}
	return pxeFiles
}
