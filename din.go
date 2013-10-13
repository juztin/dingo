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
	VERSION string = "0.1.14"
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
type Error func(ctx Context, status int) bool

var ErrorHandler Error

/*----------------------------------Handler-----------------------------------*/
type Handler func(ctx Context)

/*----------------------------------Context-----------------------------------*/
type Context struct {
	*http.Request
	Response  http.ResponseWriter
	RouteData map[string]string
}

func NewContext(response http.ResponseWriter, request *http.Request) Context {
	c := new(Context)
	c.Request, c.Response = request, response
	return *c
}

func (c *Context) Redirect(path string) {
	http.Redirect(c.Response, c.Request, path, http.StatusFound)
}

func (c *Context) RedirectPerm(path string) {
	r := c.Response
	r.Header().Set("Location", path)
	r.WriteHeader(http.StatusMovedPermanently)
}

//func (c *Context) HttpError(status int, msg ...[]byte) {
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
type Route interface {
	Path() string
	IsCanonical() bool
	Matches(path string) bool
	Execute(ctx Context)
}

/*----------------------------------Routes------------------------------------*/
type Routes interface {
	Route(url string) (Route, bool)
	Add(route Route)
}

type routes []Route

func (r *routes) Route(url string) (Route, bool) {
	for _, route := range *r {
		if route.Matches(url) {
			return route, true
		}
	}
	return nil, false
}

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

type NewIRoute func(path string, h interface{}) Route
type Router struct {
	svr   *Server
	path  string
	route NewIRoute
}

func NewRouter(s *Server, path string, route NewIRoute) Router {
	return Router{s, path, route}
}
func (r Router) add(h interface{}, m string) Router {
	route := r.route(r.path, h)
	r.svr.Route(route, m)
	return r
}
func (r Router) Options(h interface{}) Router {
	return r.add(h, "OPTIONS")
}
func (r Router) Get(h interface{}) Router {
	return r.add(h, "GET")
}
func (r Router) Post(h interface{}) Router {
	return r.add(h, "POST")
}
func (r Router) Put(h interface{}) Router {
	return r.add(h, "PUT")
}
func (r Router) Delete(h interface{}) Router {
	return r.add(h, "DELETE")
}
func (r Router) Trace(h interface{}) Router {
	return r.add(h, "TRACE")
}
func (r Router) Connect(h interface{}) Router {
	return r.add(h, "CONNECT")
}

/*-----------------------------------Server-----------------------------------*/
type Server struct {
	listener net.Listener
	routes   map[string]Routes
}

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

func New(l net.Listener) Server {
	s := new(Server)
	s.initRoutes()
	s.listener = l

	return *s
}

func (s *Server) Route(rt Route, methods ...string) {
	for _, m := range methods {
		if r, ok := s.routes[m]; ok {
			r.Add(rt)
		} else {
			log.Printf("Invalid method: %s, for route: %s\n", m, rt.Path())
		}
	}
}
func (s *Server) SRoute(path string, handler Handler, methods ...string) {
	rt := NewSRoute(path, handler)
	s.Route(rt, methods...)
}
func (s *Server) ReRoute(path string, handler Handler, methods ...string) {
	rt := NewReRoute(path, handler)
	s.Route(rt, methods...)
}
func (s *Server) RRoute(path string, handler interface{}, methods ...string) {
	rt := NewRRoute(path, handler)
	s.Route(rt, methods...)
}

func (s *Server) SRouter(p string) Router {
	return Router{s, p, iroute(NewSRoute)}
}
func (s *Server) ReRouter(p string) Router {
	return Router{s, p, iroute(NewReRoute)}
}
func (s *Server) RRouter(p string) Router {
	return Router{s, p, NewRRoute}
}

func (s *Server) Serve() {
	http.Serve(s.listener, s)
}

/* http.Handler */
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r)
	defer _500Handler(ctx)

	if routes, ok := s.routes[r.Method]; ok {
		path, canonical := IsCanonical(r.URL.Path)
		if r, ok := routes.Route(path); ok {
			if r.IsCanonical() && !canonical {
				ctx.RedirectPerm(path)
			} else {
				r.Execute(ctx)
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
