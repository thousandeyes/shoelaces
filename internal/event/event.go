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

package event

import (
	"encoding/json"
	"time"

	"github.com/thousandeyes/shoelaces/internal/server"
)

// Type holds the different typs of events
type Type int

const (
	// HostPoll is the event generated when a host poll Shoelaces for a script
	HostPoll Type = 0
	// UserSelection is the event generated when a user selects a script and hits "Boot!"
	UserSelection Type = 1
	// HostBoot is the event generated when a host finally boots
	HostBoot Type = 2
	// HostTimeout is the event generated when a host polls and after some
	// minutes without activity, timeouts.
	HostTimeout Type = 3

	// PtrMatchBoot is triggered when a PTR is matched to an IP
	PtrMatchBoot = "DNS Match"
	// SubnetMatchBoot is triggered when an IP matches a subnet mapping
	SubnetMatchBoot = "Subnet Match"
	// ManualBoot is triggered when the user selects manual boot
	ManualBoot = "Manual"
)

// Event holds information related to the interactions of hosts when they boot.
// It's used exclusively in the Shoelaces web frontend.
type Event struct {
	Type     Type                   `json:"eventType"`
	Date     time.Time              `json:"date"`
	Server   server.Server          `json:"server"`
	BootType string                 `json:"bootType"`
	Script   string                 `json:"script"`
	Message  string                 `json:"message"`
	Params   map[string]interface{} `json:"params"`
}

// Log holds the events log
type Log struct {
	Events map[string][]Event
}

// New creates a new Event object
func New(eventType Type, srv server.Server, bootType, script string, params map[string]interface{}) Event {
	var event Event

	event.Type = eventType
	event.Date = time.Now()
	event.Server = srv
	event.BootType = bootType
	event.Script = script
	event.Params = params

	event.setMessage()

	return event
}

func (e *Event) setMessage() {
	switch e.Type {
	case HostPoll:
		e.Message = "Host " + e.Server.Hostname + " polled for a script."
	case UserSelection:
		e.Message = "A user selected " + e.Script + " for the host " + e.Server.Hostname + "."
	case HostBoot:
		params, _ := json.Marshal(e.Params)
		e.Message = "Host " + e.Server.Hostname + " booted using " + e.BootType + " method with the following parameters: " + string(params)
	case HostTimeout:
		e.Message = "Host " + e.Server.Hostname + " timed out."
	}
}

// AddEvent adds an Event into the event log
func (el *Log) AddEvent(eventType Type, srv server.Server, bootType string, script string, params map[string]interface{}) {
	if el.Events == nil {
		el.Events = make(map[string][]Event)
	}

	el.Events[srv.Mac] = append(el.Events[srv.Mac], New(eventType, srv, bootType, script, params))
}
