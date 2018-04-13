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

package mappings

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/thousandeyes/shoelaces/internal/log"
)

// Mappings struct contains YamlNetworkMaps and YamlHostnameMaps.
type Mappings struct {
	NetworkMaps  []YamlNetworkMap  `yaml:"networkMaps"`
	HostnameMaps []YamlHostnameMap `yaml:"hostnameMaps"`
}

// YamlNetworkMap struct contains an association between a CIDR network and a
// Script. It's different than mapping.NetworkMap in the sense that this
// struct can be used to parse the JSON mapping file.
type YamlNetworkMap struct {
	Network string
	Script  YamlScript
}

// YamlHostnameMap struct contains an association between a hostname regular
// expression and a Script. It's different than mapping.HostnameMap in the
// sense that this struct can be used to parse the JSON mapping file.
type YamlHostnameMap struct {
	Hostname string
	Script   YamlScript
}

// YamlScript holds information regarding a script. Its name, its environment
// and its parameters.
type YamlScript struct {
	Name        string
	Environment string
	Params      map[string]string
}

// ParseYamlMappings parses the mappings yaml file into a Mappings struct.
func ParseYamlMappings(logger log.Logger, mappingsFile string) *Mappings {
	var mappings Mappings

	logger.Info("component", "config", "msg", "Reading mappings", "source", mappingsFile)
	yamlFile, err := ioutil.ReadFile(mappingsFile)

	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	mappings.NetworkMaps = make([]YamlNetworkMap, 0)
	mappings.HostnameMaps = make([]YamlHostnameMap, 0)

	err = yaml.Unmarshal(yamlFile, &mappings)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	return &mappings
}
