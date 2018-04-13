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

package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
)

// StaticConfigFileHandler handles static config files
type StaticConfigFileHandler struct{}

func (s *StaticConfigFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	env := envFromRequest(r)
	envName := envNameFromRequest(r)
	basePath := path.Join(env.DataDir, "static")
	if envName == "" {
		http.FileServer(http.Dir(basePath)).ServeHTTP(w, r)
		return
	}
	envPath := filepath.Join(env.DataDir, env.EnvDir, envName, "static")
	OverlayFileServer(envPath, basePath).ServeHTTP(w, r)
}

// StaticConfigFileServer returns a StaticConfigFileHandler instance implementing http.Handler
func StaticConfigFileServer() *StaticConfigFileHandler {
	return &StaticConfigFileHandler{}
}

// OverlayFileServerHandler handles request for overlayer directories
type OverlayFileServerHandler struct {
	upper string
	lower string
}

// OverlayFileServer serves static content from two overlayed directories
func OverlayFileServer(upper, lower string) *OverlayFileServerHandler {
	return &OverlayFileServerHandler{
		upper: upper,
		lower: lower,
	}
}

func (o *OverlayFileServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fp := filepath.Clean(r.URL.Path)
	upper := filepath.Clean(path.Join(o.upper, fp))
	lower := filepath.Clean(path.Join(o.lower, fp))

	// TODO: try to avoid stat()-ing both if not necessary
	infoUpper, errUpper := os.Stat(upper)
	infoLower, errLower := os.Stat(lower)

	// If both upper and lower files/dirs do not exist, return 404
	if errUpper != nil && os.IsNotExist(errUpper) &&
		errLower != nil && os.IsNotExist(errLower) {
		http.NotFound(w, r)
		return
	}

	isDir := false
	fileList := make(map[string]os.FileInfo)

	if errUpper == nil && infoUpper.IsDir() {
		files, _ := ioutil.ReadDir(upper)
		for _, f := range files {
			fileList[f.Name()] = f
		}
		isDir = true
	}
	if errLower == nil && infoLower.IsDir() {
		files, _ := ioutil.ReadDir(lower)
		for _, f := range files {
			if _, ok := fileList[f.Name()]; !ok {
				fileList[f.Name()] = f
			}
		}
		isDir = true
	}

	// Generate HTML directory index
	if isDir {
		fileListIndex := []string{}
		for i := range fileList {
			fileListIndex = append(fileListIndex, i)
		}
		sort.Strings(fileListIndex)
		w.Write([]byte("<pre>\n"))
		for _, i := range fileListIndex {
			f := fileList[i]
			name := f.Name()
			if f.IsDir() {
				name = name + "/"
			}
			l := fmt.Sprintf("<a href=\"%s\">%s</a>\n", name, name)
			w.Write([]byte(l))
		}
		w.Write([]byte("</pre>\n"))
		return
	}

	// Serve the file from the upper layer if it exists.
	if errUpper == nil {
		http.ServeFile(w, r, upper)
		// If not serve it from the lower
	} else if errLower == nil {
		http.ServeFile(w, r, lower)
	}
	http.NotFound(w, r)
}
