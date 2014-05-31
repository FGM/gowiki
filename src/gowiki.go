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
	"regexp"
)

const PORT = 8080

type Page struct {
	Title string
	Body  []byte
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := loadPage(title)
	// S'il n'y a pas de page, en créer une vierge avec le titre.
	if err != nil {
		page = &Page{Title: title}
	}

	renderTemplate(w, "edit", page)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid page title")
	}
	// 0: all, 1: op, 2: title.
	return m[2], nil
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	ret := &Page{Title: title, Body: body}
	return ret, nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
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
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", page)
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	ret := ioutil.WriteFile(filename, p.Body, os.FileMode(0600))
	return ret
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([_a-zA-Z0-9]+)$")

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	listenAddress := fmt.Sprintf(":%d", PORT)
	fmt.Printf("Listening on %s\n", listenAddress)
	http.ListenAndServe(listenAddress, nil)
}
