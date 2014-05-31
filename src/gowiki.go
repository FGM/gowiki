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
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
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

// ---- Configuration ----------------------------------------------------------
// Derived configuration globals.
// @see var conf
type Configuration struct {
	Templates        *template.Template
	ValidPath        *regexp.Regexp
}

func (c *Configuration) init(s Settings) {
	c.Templates = template.Must(template.ParseGlob(s.TemplatesPath + "/*"))
	c.ValidPath = regexp.MustCompile("^/((edit|save|view)/([_a-zA-Z0-9]+))?$")
}

// ---- Page -------------------------------------------------------------------
// A wiki page.
type Page struct {
	Title string
	Body  []byte
}

// Load page with given title.
// If page file does not exist, set Title and leave Body empty.
func (p *Page) load(title string) {
	filename := settings.PagesPath + "/" + title + ".txt"
	p.Title = title
	// A failed read is normal at this point, just leave Body blank.
	p.Body, _ = ioutil.ReadFile(filename)
}

// Save page to file named "(p.Title).txt".
func (p *Page) save() error {
	filename := settings.PagesPath + "/" + p.Title + ".txt"
	ret := ioutil.WriteFile(filename, p.Body, os.FileMode(0600))
	return ret
}

// ---- Request handler --------------------------------------------------------
type RequestHandler func(http.ResponseWriter, *http.Request, string)

// ---- Routing map ------------------------------------------------------------
type RoutingMap map[string]RequestHandler

// List paths in the routing map and their handlers.
func (m RoutingMap) dump() int {
	for path, handler := range m {
		fmt.Println(path, reflect.TypeOf(handler), handler)
	}
	return len(m)
}

// Register the routes in the http dispatcher.
func (m RoutingMap) register() {
	for path, handler := range m {
		http.HandleFunc(path, makeHandler(handler))
	}
}

// ---- Settings ---------------------------------------------------------------
// Initial configuration values.
type Settings struct {
	Port             uint16
	PagesPath        string
	TemplatesPath    string
	Routes			 *RoutingMap
}

// ---- Functions --------------------------------------------------------------
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

// ---- Main code --------------------------------------------------------------

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
