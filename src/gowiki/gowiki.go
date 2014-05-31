/**
 * Web programming in Go: tutorial example.
 *
 * http://golang.org/doc/articles/wiki/
 */

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"net/http"
)

// Environment related initial configuration.
var settings = Settings{
	Port:          8080,
	TemplatesPath: "templates",
	PagesPath:     "data",
	Routes: &RoutingMap{
		"/"	     : frontHandler,
		"/view/" : viewHandler,
		"/edit/" : editHandler,
		"/save/" : saveHandler,
	},
}

type RequestHandler func(http.ResponseWriter, *http.Request, string)


// Handler for edit/* pages.
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	var page Page

	page.load(title)
	fmt.Println("Editing", title)
	renderTemplate(w, "edit", &page)
}

// Handler for front page.
func frontHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("Home redirecting to FrontPage")
    http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := conf.ValidPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid page title")
	}
	// 0: all, 1: op, 2: title.
	return m[2], nil
}

// Build a handler acceptable by the http dispatcher from a RequestHandler.
func makeHandler(fn RequestHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := conf.ValidPath.FindStringSubmatch(r.URL.Path)
		// fmt.Printf("URL: %s, Matches %#v, len %d\n", r.URL.Path, m, len(m))
		if m == nil {
			http.NotFound(w, r)
			return
		}
		// On normal paths: 0: all, 1: op, 2: title.
		// len(m) is always 4, so we do not need a special case for "/".
        fn(w, r, m[3])
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := conf.Templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("Saving", title)
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}

	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("Viewing", title)
	var page Page

	page.load(title)
	if len(page.Body) == 0 {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", &page)
}

// A single wrapper for all application-defined globals, derived from settings.
var conf Configuration

// Initialize computed globals from data at the top of the file, and register
// route handlers.
func init() {
	conf.init(settings)
	settings.Routes.register()
	// settings.Routes.dump()
}

// Main function: HTTP server.
func main() {
	listenAddress := fmt.Sprintf(":%d", settings.Port)
	fmt.Printf("Listening on %s\n", listenAddress)
	http.ListenAndServe(listenAddress, nil)
}
