// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dingo is a simple wrapper around net/http providing very
// simple routing, with RegEx patterns, and views (templates).
package dingo

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"runtime"
)

const (
	VERSION string = "0.1.15"
)

var (
	httpMethods = []string{
		"OPTIONS",
		"GET",
		"HEAD",
		"POST",
		"PUT",
		"DELETE",
		"TRACE",
		"CONNECT",
	}
)

/*-----------------------------------Error------------------------------------*/

// Error func is a handler type called when there is a panic.
type Error func(ctx Context, status int) bool

// ErrorHandler is the actual handler invoked upon panic.
var ErrorHandler Error

/*----------------------------------Handler-----------------------------------*/

// Handler is a func type called when a matching route is found.
type Handler func(ctx Context)

/*----------------------------------Context-----------------------------------*/

// Context holds both the request and response objects, along with route-data,
// and is passed to handlers.
type Context struct {
	*http.Request
	Response  http.ResponseWriter
	RouteData map[string]string
}

// NewContext creates, and returns, a new Context
func NewContext(response http.ResponseWriter, request *http.Request) Context {
	c := new(Context)
	c.Request, c.Response = request, response
	return *c
}

// Redirect issues a redirect to the response.
func (c *Context) Redirect(path string) {
	http.Redirect(c.Response, c.Request, path, http.StatusFound)
}

// RedirectPerm issues a permanent redirect to the response.
func (c *Context) RedirectPerm(path string) {
	r := c.Response
	r.Header().Set("Location", path)
	r.WriteHeader(http.StatusMovedPermanently)
}

// HttpError issues the given http error (status) and writes any provided msgs.
// The ErrorHandler is invoked when it has been set and there are no provided msgs.
func (c *Context) HttpError(status int, msgs ...string) {
	//if ErrorHandler != nil {
	if ErrorHandler != nil && msgs == nil {
		// should we defer a panic here? Or just assume you know what you're doing?
		if ErrorHandler(*c, status) {
			return
		}
	}

	// default to 500 if in invalid status is given
	if status < 100 || status > 505 {
		status = 500
		// TODO Log an error
	}
	r := c.Response
	r.WriteHeader(status)
	//if msgs == nil {
	if len(msgs) == 0 {
		m := []byte(http.StatusText(status))
		r.Write(m)
	} else {
		for _, msg := range msgs {
			r.Write([]byte(msg))
		}
	}
}

/*----------------------------------Route-------------------------------------*/

// Route is the handling object when a request matches.
type Route interface {
	Path() string
	IsCanonical() bool
	Matches(path string) bool
	Execute(ctx Context)
}

/*----------------------------------Routes------------------------------------*/

// Routes is a collection of Route types.
type Routes interface {
	Route(url string) (Route, bool)
	Add(route Route)
}

type routes []Route

// Route finds a Route for the given url, or nil.
func (r *routes) Route(path string) (Route, bool) {
	for _, route := range *r {
		if route.Matches(path) {
			return route, true
		}
	}
	return nil, false
}

// Add adds a new route to the collection.
func (r *routes) Add(route Route) {
	*r = append(*r, route)
}

/*----------------------------------Router------------------------------------*/

type newRoute func(path string, h Handler) Route

func iroute(route newRoute) NewIRoute {
	var fn NewIRoute
	fn = func(path string, h interface{}) Route {
		switch h.(type) {
		case Handler:
			return route(path, h.(Handler))
		case func(Context):
			return route(path, Handler(h.(func(Context))))
		}

		panic(fmt.Sprintf("Handler is invalid: %v", h))
	}
	return fn
}

// NewIRoute returns a new route for the given path and handler.
type NewIRoute func(path string, h interface{}) Route
type Router struct {
	svr   *Server
	path  string
	route NewIRoute
}

// NewRouter returns a router.
func NewRouter(s *Server, path string, route NewIRoute) Router {
	return Router{s, path, route}
}
func (r Router) add(h interface{}, m string) Router {
	route := r.route(r.path, h)
	r.svr.Route(route, m)
	return r
}

// Options adds a route for "OPTIONS" method.
func (r Router) Options(h interface{}) Router {
	return r.add(h, "OPTIONS")
}

// Options adds a route for "GET" method.
func (r Router) Get(h interface{}) Router {
	return r.add(h, "GET")
}

// Options adds a route for "POST" method.
func (r Router) Post(h interface{}) Router {
	return r.add(h, "POST")
}

// Options adds a route for "PUT" method.
func (r Router) Put(h interface{}) Router {
	return r.add(h, "PUT")
}

// Options adds a route for "DELETE" method.
func (r Router) Delete(h interface{}) Router {
	return r.add(h, "DELETE")
}

// Options adds a route for "TRACE" method.
func (r Router) Trace(h interface{}) Router {
	return r.add(h, "TRACE")
}

// Options adds a route for "CONNECT" method.
func (r Router) Connect(h interface{}) Router {
	return r.add(h, "CONNECT")
}

/*-----------------------------------Server-----------------------------------*/

// Server implements http.Handler and routes calls to handlers view a Routes collection.
type Server struct {
	listener net.Listener
	routes   map[string]Routes
}

// IsCanonical returns wether the given path is canonical.
func IsCanonical(p string) (string, bool) {
	if len(p) == 0 {
		return "/", false
	} else if p[0] != '/' {
		return "/" + p, false
	}

	cp := path.Clean(p)

	if cp[len(cp)-1] != '/' {
		cp = cp + "/"
		return cp, cp == p
	}

	return cp, cp == p
}

func (s *Server) initRoutes() {
	s.routes = make(map[string]Routes)
	for _, method := range httpMethods {
		s.routes[method] = new(routes)
	}
}

// New returns a new Server.
func New(l net.Listener) Server {
	s := new(Server)
	s.initRoutes()
	s.listener = l

	return *s
}

// Route adds a route for the given methods.
func (s *Server) Route(rt Route, methods ...string) {
	for _, m := range methods {
		if r, ok := s.routes[m]; ok {
			r.Add(rt)
		} else {
			log.Printf("Invalid method: %s, for route: %s\n", m, rt.Path())
		}
	}
}

// SRoute adds a new static route.
func (s *Server) SRoute(path string, handler Handler, methods ...string) {
	rt := NewSRoute(path, handler)
	s.Route(rt, methods...)
}

// ReRoute adds a new RegEx route.
func (s *Server) ReRoute(path string, handler Handler, methods ...string) {
	rt := NewReRoute(path, handler)
	s.Route(rt, methods...)
}

// RRoute adds a new RegEx route where the params are passed as params to the given handler.
func (s *Server) RRoute(path string, handler interface{}, methods ...string) {
	rt := NewRRoute(path, handler)
	s.Route(rt, methods...)
}

// SRouter returns a new Router used to add static routes to the server.
func (s *Server) SRouter(p string) Router {
	return Router{s, p, iroute(NewSRoute)}
}

// SRouter returns a new Router used to add RegEx routes to the server.
func (s *Server) ReRouter(p string) Router {
	return Router{s, p, iroute(NewReRoute)}
}

// SRouter returns a new Router used to add RegEx routes to the server where the params are passed to the handler.
func (s *Server) RRouter(p string) Router {
	return Router{s, p, NewRRoute}
}

// Serve begins listening for requests
func (s *Server) Serve() {
	http.Serve(s.listener, s)
}

/* http.Handler */
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r)
	defer _500Handler(ctx)

	if routes, ok := s.routes[r.Method]; ok {
		path, canonical := IsCanonical(r.URL.Path)
		if rt, ok := routes.Route(path); ok {
			if rt.IsCanonical() && !canonical {
				r.URL.Path = path
				ctx.RedirectPerm(r.URL.String())
			} else {
				rt.Execute(ctx)
			}
			return
		}
	}

	ctx.HttpError(404)
}

/*--------------*/

func _500Handler(ctx Context) {
	//if err, ok := recover().(error); ok {
	if err := recover(); err != nil {
		// TODO write the 500 message, or stack, depending on some settings
		// if !emitError
		ctx.HttpError(500)

		log.Println("_______________________________________ERR______________________________________")
		log.Println(err)
		for i := 1; ; i++ {
			if _, f, l, ok := runtime.Caller(i); !ok {
				break
			} else {
				//p := strings.Split(f, "/")
				//fmt.Printf("%s : %d\n", p[len(p)-1], l)
				//fmt.Printf("%s\n%d\n_________________________\n", f, l)
				log.Printf("Line: %d\nfile: %s\n-\n", l, f)
			}
		}
		log.Println("________________________________________________________________________________")
	}
}
