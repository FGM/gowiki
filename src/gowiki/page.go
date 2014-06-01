package main

import (
	"io/ioutil"
	"os"
)

/*
 	Page represents a wiki page, as loaded in program memory.

 As such, it contains:

  * the page title, used for the URL
  * the page body, as an array of bytes instead of a string, to ease filesystem
    input/output

  */
type Page struct {
	Title string
	Body  []byte
}

// Load page with given title.
// If page file does not exist, set Title and leave Body empty.
func (p *Page) load(title string) {
	filename := Conf.Settings.PagesPath + "/" + title + ".txt"
	p.Title = title
	// A failed read is normal at this point, just leave Body blank.
	p.Body, _ = ioutil.ReadFile(filename)
}

// Save page to file named "(p.Title).txt".
func (p *Page) save() error {
	filename := Conf.Settings.PagesPath + "/" + p.Title + ".txt"
	ret := ioutil.WriteFile(filename, p.Body, os.FileMode(0600))
	return ret
}
