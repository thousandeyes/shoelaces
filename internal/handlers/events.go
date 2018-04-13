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
	"encoding/json"
	"net/http"
	"os"
)

// ListEvents returns a JSON list of the logged events.
func ListEvents(w http.ResponseWriter, r *http.Request) {
	// Get Environment and convert the EventLog to JSON
	env := envFromRequest(r)
	eventList, err := json.Marshal(env.EventLog.Events)
	if err != nil {
		env.Logger.Error("component", "handler", "err", err)
		os.Exit(1)
	}

	//Write the EventLog and send the HTTP response
	w.Header().Set("Content-Type", "application/json")
	w.Write(eventList)
}
