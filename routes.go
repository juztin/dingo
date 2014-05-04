// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dingo

import (
	"reflect"
	"regexp"
)

/*----------static----------*/

type route struct {
	path    string
	handler Handler
}

// NewSRoute returns a new static route for the given path/handler.
func NewSRoute(path string, handler Handler) Route {
	rt := new(route)
	rt.path = path
	rt.handler = handler

	//return *rt
	return rt
}

// Returns the path of the route.
func (r *route) Path() string {
	return r.path
}

// IsCanonical returns if the route is canonical.
func (r *route) IsCanonical() bool {
	return true
}

// Matches returns wether this route matches the given path.
func (r *route) Matches(path string) bool {
	return r.path == path
}

// Execute invokes the handler for this route, writing it's response.
func (r *route) Execute(ctx Context) {
	r.handler(ctx)
}

/*----------regexp----------*/

type reRoute struct {
	path    string
	expr    *regexp.Regexp
	handler Handler
}

// NewReRoute returns a new RegEx route for the given path/handler.
func NewReRoute(re string, handler Handler) Route {
	rt := new(reRoute)
	rt.path = re
	rt.expr = regexp.MustCompile(re)
	rt.handler = handler

	return rt
}

func (r *reRoute) data(url string) map[string]string {
	data := make(map[string]string)
	//matches := r.expr.FindAllStringSubmatch(url, -1)
	matches := r.expr.FindStringSubmatch(url)

	for i, n := range r.expr.SubexpNames() {
		if i == 0 {
			continue
		}
		data[n] = matches[i]
	}

	return data
}

// Returns the path of the route.
func (r *reRoute) Path() string {
	return r.path
}

// IsCanonical returns if the route is canonical.
func (r *reRoute) IsCanonical() bool {
	return true
}

// Matches returns wether this route matches the given path.
func (r *reRoute) Matches(url string) bool {
	return r.expr.MatchString(url)
}

// Execute invokes the handler for this route, writing it's response.
func (r *reRoute) Execute(ctx Context) {
	ctx.RouteData = r.data(ctx.URL.Path)
	r.handler(ctx)
}

/*---------regexp2----------*/

type rRoute struct {
	path    string
	expr    *regexp.Regexp
	handler reflect.Value
}

// NewRRoute returns a new RegEx route for the given path/handler.
func NewRRoute(re string, handler interface{}) Route {
	r := new(rRoute)
	r.path = re
	r.expr = regexp.MustCompile(re)

	if fn, ok := handler.(reflect.Value); ok {
		r.handler = fn
	} else {
		r.handler = reflect.ValueOf(handler)
	}

	return r
}

// Returns the path of the route.
func (r *rRoute) Path() string {
	return r.path
}

// IsCanonical returns if the route is canonical.
func (r *rRoute) IsCanonical() bool {
	return true
}

// Matches returns wether this route matches the given path.
func (r *rRoute) Matches(url string) bool {
	return r.expr.MatchString(url)
}

// Execute invokes the handler, passing in the RegEx groups, for this route, writing it's response.
func (r *rRoute) Execute(ctx Context) {
	// TODO it would be nice if we could detect numbers and cast them as such prior to invoking the func
	args := []reflect.Value{reflect.ValueOf(ctx)}
	matches := r.expr.FindStringSubmatch(ctx.URL.Path)
	for _, a := range matches[1:] {
		args = append(args, reflect.ValueOf(a))
	}
	r.handler.Call(args)
}
