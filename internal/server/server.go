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

package server

import (
	"sync"
	"time"

	"github.com/thousandeyes/shoelaces/internal/log"
)

const (
	// InitTarget is an initial dummy target assigned to the servers
	InitTarget = "NOTARGET"
)

// Server holds data that uniquely identifies a server
type Server struct {
	Mac      string
	IP       string
	Hostname string
}

// Servers is an array of Server
type Servers []Server

// Len implementation for the sort Interface
func (s Servers) Len() int {
	return len(s)
}

// Swap implementation for the sort interface
func (s Servers) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implementation for the Sort interface
func (s Servers) Less(i, j int) bool {
	return s[i].Mac < s[j].Mac
}

// State holds information regarding a host that is attempting to boot.
type State struct {
	Server
	Target      string
	Environment string
	Params      map[string]interface{}
	Retry       int
	LastAccess  int
}

// States holds a map between MAC addresses and
// States. It provides a mutex for thread-safety.
type States struct {
	sync.RWMutex
	Servers map[string]*State
}

// New returns a Server with is values initialized
func New(mac string, ip string, hostname string) Server {
	return Server{
		Mac:      mac,
		IP:       ip,
		Hostname: hostname,
	}
}

// AddServer adds a server to the States struct
func (m *States) AddServer(server Server) {
	m.Servers[server.Mac] = &State{
		Server:     server,
		Target:     InitTarget,
		Retry:      1,
		LastAccess: int(time.Now().UTC().Unix()),
	}
}

// DeleteServer deletes a server from the States struct
func (m *States) DeleteServer(mac string) {
	delete(m.Servers, mac)
}

// StartStateCleaner spawns a goroutine that cleans MAC addresses that
// have been inactive in Shoelaces for more than 3 minutes.
func StartStateCleaner(logger log.Logger, serverStates *States) {
	const (
		// 3 minutes
		expireAfterSec = 3 * 60
		cleanInterval  = time.Minute
	)
	// Clean up the server states. Expire after 3 minutes
	go func() {
		for {
			time.Sleep(cleanInterval)

			servers := serverStates.Servers
			expire := int(time.Now().UTC().Unix()) - expireAfterSec

			logger.Debug("component", "polling", "msg", "Cleaning", "before", time.Unix(int64(expire), 0))

			serverStates.Lock()
			for mac, state := range servers {
				if state.LastAccess <= expire {
					delete(servers, mac)
					logger.Debug("component", "polling", "msg", "Mac cleaned", "mac", mac)
				}
			}
			serverStates.Unlock()
		}
	}()
}
