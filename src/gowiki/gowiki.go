/*
Web programming in Go: tutorial example.

  Initial commit	based on code from http://golang.org/doc/articles/wiki/
			Copyright 2010 The Go Authors. All rights reserved.
			Use of this source code is governed by a BSD-style
			license that can be found in the LICENSE file.

  Later commits		non-reference newbie code
			Copyright 2014 OSInet. All rights reserved.
			Use of this source code is governed by a BSD-style
			license that can be found in the LICENSE file.

*/
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RequestHandler func(http.ResponseWriter, *http.Request, string)

// EditHandler handles edit/* pages.
//
// TODO(routing) remove "r" parameter when routing will permit it.
func EditHandler(w http.ResponseWriter, r *http.Request, title string) {
	var page Page

	page.load(title)
	fmt.Println("Editing", title)
	RenderTemplate(w, "edit", &page)
}

// FrontHandler handles front page.
//
// TODO(routing) remove "title" parameter when routing will permit it.
func FrontHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("Home redirecting to FrontPage")
    http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

// GetTitle extracts the page title from a wiki action page (view/edit/save).
func GetTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := Conf.ValidPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid page title")
	}
	// 0: all, 1: op, 2: title.
	return m[2], nil
}

// MakeHandler builds a handler acceptable by the http dispatcher from a RequestHandler.
func MakeHandler(fn RequestHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := Conf.ValidPath.FindStringSubmatch(r.URL.Path)
		// fmt.Printf("URL: %s, Matches %#v, len %d, path %s\n", r.URL.Path, m, len(m), m[3])
		if m == nil {
			http.NotFound(w, r)
			return
		}
		// On normal paths: 0: all, 1: op, 2: title.
		// len(m) is always 4, so we do not need a special case for "/".
        fn(w, r, m[3])
	}
}

/*
RenderTemplate renders a page template.

Arguments

	w	A ResponseWriter on which to write results
	tmpl	The base name of the template, without ".html"
	p	The page to render
*/
func RenderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	vars := make(map[string]interface {})

	vars["page"] = p
	vars["settings"] = Conf.Settings

	err := Conf.Templates.ExecuteTemplate(w, tmpl+".html", vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

/*
SaveHandler handles "save" pages.

It saves the file for the page, overwriting any previous version, and redirects
to the view for the same page.

Arguments

	w	A ResponseWriter on which to write results
	r	The Request
	title	The title of the page to save
 */
func SaveHandler(w http.ResponseWriter, r *http.Request, title string) {
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

/*
StylesHandler handles loading CSS files from /styles/*

Arguments

	w	A ResponseWriter on which to write results
	r	The Request
	title	Ignored parameter, only made necessary by current routing code.

TODO(routing) remove "title" parameter when routing will permit it.
*/
func StylesHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("CSS:", title)

	filename := Conf.Settings.StylesPath
	css, err := ioutil.ReadFile(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write(css)
}

/*
ViewHandler handles the "view" pages.

Arguments

	w	A ResponseWriter on which to write results
	r	The Request
	title	The title of the page to view
 */
func ViewHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("Viewing", title)
	var page Page

	page.load(title)
	if len(page.Body) == 0 {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	RenderTemplate(w, "view", &page)
}

// A single wrapper for all application-defined globals, derived from settings.
var Conf Configuration

// Initialize computed globals from data at the top of the file, and register
// route handlers.
func Init() {
	var settings = Settings{
		Port:          8080,
		StylesPath:	   "styles/wiki.css",
		TemplatesPath: "templates",
		PagesPath:     "data",
		Routes: &RoutingMap{
			"/"	      : FrontHandler,
			"/view/"  : ViewHandler,
			"/edit/"  : EditHandler,
			"/save/"  : SaveHandler,
			"/styles/": StylesHandler,
		},
	}

	settings.Routes.register()
	Conf.init(settings)
	settings.Routes.dump()
}

/*
Main function: HTTP server.

	- Listen and serve on port configured in Conf.settings.
	- Assumes configuration has been loaded by init()/Init().
  */
func Main() {
	listenAddress := fmt.Sprintf(":%d", Conf.Settings.Port)
	fmt.Printf("Listening on %s\n", listenAddress)
	http.ListenAndServe(listenAddress, nil)
}

// Go needs these as private functions to work.
func init() { Init() }
func main() { Main() }
