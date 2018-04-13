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
	"testing"
	"time"

	"github.com/thousandeyes/shoelaces/internal/server"
)

const expectedEvent = `{"eventType":0,"date":"1970-01-01T00:00:00Z","server":{"Mac":"","IP":"","Hostname":"test_host"},"bootType":"Manual","script":"freebsd.ipxe","message":"","params":{"baseURL":"localhost:8080","cloudconfig":"virtual","hostname":"","version":"12345"}}`

func TestNew(t *testing.T) {
	event := New(HostPoll, server.Server{Mac: "", IP: "", Hostname: "test_host"}, PtrMatchBoot, "msdos.ipxe", map[string]interface{}{"test": "testParam"})
	if event.Type != HostPoll {
		t.Errorf("Expected: \"%d\"\nGot: \"%d\"", HostPoll, event.Type)
	}
	if event.Server.Hostname != "test_host" {
		t.Errorf("Expected: \"test_host\"\nGot: \"%s\"", event.Server.Hostname)
	}
	if event.BootType != PtrMatchBoot {
		t.Errorf("Expected: \"%s\"\nGot: \"%s\"", PtrMatchBoot, event.Server.Hostname)
	}
	if event.Script != "msdos.ipxe" {
		t.Errorf("Expected: \"msdos.ipxe\"\nGot: \"%s\"", event.Server.Hostname)
	}
	if len(event.Params) != 1 {
		t.Error("Expected one parameter")
	}
	if event.Params["test"] != "testParam" {
		t.Error("Expected parameter test: testParam")
	}
	now := time.Now()
	if event.Date.After(now) {
		t.Errorf("Expected %s to be after %s", event.Date, now)
	}
}

func TestEventMarshalJSON(t *testing.T) {
	event := Event{
		Type:     HostPoll,
		Date:     time.Unix(0, 0).UTC(),
		Server:   server.Server{Mac: "", IP: "", Hostname: "test_host"},
		BootType: ManualBoot,
		Script:   "freebsd.ipxe",
		Message:  "",
		Params: map[string]interface{}{
			"baseURL":     "localhost:8080",
			"cloudconfig": "virtual",
			"hostname":    "",
			"version":     "12345",
		},
	}
	marshaled, _ := json.Marshal(event)
	if string(marshaled) != expectedEvent {
		t.Errorf("Expected %s\nGot: %s\n", expectedEvent, marshaled)
	}
}
