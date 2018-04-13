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
	"testing"

	"github.com/thousandeyes/shoelaces/internal/mappings"
)

func TestDefaultEnvironment(t *testing.T) {
	env := defaultEnvironment()
	if env.BaseURL != "" {
		t.Error("BaseURL should be empty string if instantiated directly.")
	}
	if len(env.HostnameMaps) != 0 {
		t.Error("Hostname mappings should be empty")
	}
	if len(env.NetworkMaps) != 0 {
		t.Error("Network mappings should be empty")
	}
	if len(env.ParamsBlacklist) != 1 &&
		env.ParamsBlacklist[0] != "baseURL" {
		t.Error("ParamsBlacklist should have only baseURL")
	}
}

func TestInitScript(t *testing.T) {
	params := make(map[string]string)
	params["one"] = "one_value"
	configScript := mappings.YamlScript{Name: "testscript", Params: params}
	mappingScript := initScript(configScript)
	if mappingScript.Name != "testscript" {
		t.Errorf("Expected: %s\nGot: %s\n", "testscript", mappingScript.Name)
	}
	val, ok := mappingScript.Params["one"]
	if !ok {
		t.Error("Missing param")
	} else {
		v, ok := val.(string)
		if !ok {
			t.Error("Bad value type")
		} else {
			if v != "one_value" {
				t.Error("Bad value")
			}
		}
	}
}
