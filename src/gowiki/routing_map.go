package main

import (
	"fmt"
	"net/http"
	"reflect"
)

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


