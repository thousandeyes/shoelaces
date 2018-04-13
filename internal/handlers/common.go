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
	"html/template"
	"net/http"

	"github.com/thousandeyes/shoelaces/internal/environment"
	"github.com/thousandeyes/shoelaces/internal/ipxe"
	"github.com/thousandeyes/shoelaces/internal/mappings"
)

// DefaultTemplateRenderer holds information for rendering a template based
// on its name. It implements the http.Handler interface.
type DefaultTemplateRenderer struct {
	templateName string
}

// RenderDefaultTemplate renders a template by the given name
func RenderDefaultTemplate(name string) *DefaultTemplateRenderer {
	return &DefaultTemplateRenderer{templateName: name}
}

func (t *DefaultTemplateRenderer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	env := envFromRequest(r)
	tpl := env.StaticTemplates
	// XXX: Probably not ideal as it's doing the directory listing on every request
	ipxeScripts := ipxe.ScriptList(env)
	tplVars := struct {
		BaseURL      string
		HostnameMaps *[]mappings.HostnameMap
		NetworkMaps  *[]mappings.NetworkMap
		Scripts      *[]ipxe.Script
	}{
		env.BaseURL,
		&env.HostnameMaps,
		&env.NetworkMaps,
		&ipxeScripts,
	}
	renderTemplate(w, tpl, "header", tplVars)
	renderTemplate(w, tpl, t.templateName, tplVars)
	renderTemplate(w, tpl, "footer", tplVars)
}

func renderTemplate(w http.ResponseWriter, tpl *template.Template, tmpl string, d interface{}) {
	err := tpl.ExecuteTemplate(w, tmpl, d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func envFromRequest(r *http.Request) *environment.Environment {
	return r.Context().Value(ShoelacesEnvCtxID).(*environment.Environment)
}

func envNameFromRequest(r *http.Request) string {
	e := r.Context().Value(ShoelacesEnvNameCtxID)
	if e != nil {
		return e.(string)
	}
	return ""
}
