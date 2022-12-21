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
	"bytes"
	"fmt"
	"net/http"

	"github.com/thousandeyes/shoelaces/internal/ipxe"
)

const menuHeader = "#!ipxe\n" +
	"menu Choose target to boot\n"

const menuFooter = "\n" +
	"choose target\n" +
	"echo -n Enter hostname or none:\n" +
	"read hostname\n" +
	"set baseurl %s\n" +
	"# Boot it as intended.\n" +
	"chain ${target}\n"

// IPXEMenu serves the ipxe menu with list of all available scripts
func IPXEMenu(w http.ResponseWriter, r *http.Request) {
	env := envFromRequest(r)

	scripts := ipxe.ScriptList(env)
	if len(scripts) == 0 {
		http.Error(w, "No Scripts Found", http.StatusInternalServerError)
		return
	}

	var bootItemsBuffer bytes.Buffer
	//Creates the top portion of the iPXE menu
	bootItemsBuffer.WriteString(menuHeader)
	for _, s := range scripts {
		//Formats the bootable scripts separated by newlines into a single string
		var desc string
		if len(s.Env) > 0 {
			desc = fmt.Sprintf("%s [%s]", s.Name, s.Env)
		} else {
			desc = string(s.Name)
		}
		bootItem := fmt.Sprintf("item %s%s %s\n", s.Path, s.Name, desc)
		bootItemsBuffer.WriteString(bootItem)
	}
	//Creates the bottom portion of the iPXE menu
	bootItemsBuffer.WriteString(fmt.Sprintf(menuFooter, env.BaseURL))
	w.Write(bootItemsBuffer.Bytes())
}
