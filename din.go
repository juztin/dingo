package dingo

import (
	"fmt"
	//"io"
	"net"
	"net/http"
	"path"
)

const (
	VERSION string = "0.1.0"
)

var (
	httpMethods = []string{
		"OPTIONS", "GET", "HEAD",
		"POST", "PUT", "DELETE",
		"TRACE", "CONNECT"}
)

/*----------------------------------Handler-----------------------------------*/
type Handler func(ctx Context)

/*----------------------------------Context-----------------------------------*/
type Context struct {
	*http.Request
	Writer    http.ResponseWriter
	RouteData map[string]string
}

func newContext(request *http.Request, writer http.ResponseWriter) Context {
	c := new(Context)
	c.Request, c.Writer = request, writer
	return *c
}

func (c *Context) write(content string) {
	b := []byte(content)
	c.Writer.Write(b)
}

func (c *Context) Redirect(path string) {
	http.Redirect(c.Writer, c.Request, path, http.StatusTemporaryRedirect)
}

func (c *Context) RedirectPerm(path string) {
	w := c.Writer
	w.Header().Set("Location", path)
	w.WriteHeader(http.StatusMovedPermanently)
}

func (self *Context) HttpError(code int) {
}

/*----------------------------------Route-------------------------------------*/
type Route interface {
	Path() string
	Matches(path string) bool
	Execute(ctx Context)
}

/*----------------------------------Routes------------------------------------*/
type Routes interface {
	Route(url string) (Route, bool)
	Add(route Route)
}
type routes struct {
	routes []Route
}

func (r *routes) Route(url string) (Route, bool) {
	for _, route := range r.routes {
		if route.Matches(url) {
			return route, true
		}
	}
	return nil, false
}

func (r *routes) Add(route Route) {
	r.routes = append(r.routes, route)
}

/*-----------------------------------Server-----------------------------------*/
type Server struct {
	ip     string
	port   int
	routes map[string]Routes
	ctx    Context
}

func isCanonical(p string) (string, bool) {
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
		s.routes[method] = &routes{}
	}
}

func (s *Server) Route(rt Route, methods ...string) {
	for _, m := range methods {
		if r, ok := s.routes[m]; ok {
			r.Add(rt)
		} else {
			fmt.Printf("Invalid method: %s, for route: %s\n", m, rt.Path())
		}
	}
}
func (s *Server) StaticRoute(path string, handler Handler, methods ...string) {
	rt := NewRoute(path, handler)
	s.Route(rt, methods...)
}
func (s *Server) ReRoute(path string, handler Handler, methods ...string) {
	rt := NewReRoute(path, handler)
	s.Route(rt, methods...)
}

func (s *Server) Get(path string, handler Handler) {
	//self.routes["GET"].Add(NewRoute(path, handler))
	s.routes["GET"].Add(NewReRoute(path, handler))
}
func (s *Server) Post(path string, handler Handler) {
	s.routes["POST"].Add(NewReRoute(path, handler))
}

/* http.Handler */
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.ctx = newContext(r, w)
	defer _500Handler(s.ctx)

	if path, ok := isCanonical(r.URL.Path); !ok {
		s.ctx.RedirectPerm(path)
		return
	}

	if routes, ok := s.routes[r.Method]; ok {
		if r, ok := routes.Route(r.URL.Path); ok {
			r.Execute(s.ctx)
			return
		}
	}

	//ctx.Render("404", nil)
	s.ctx.write("404")
}

/*--------------*/
func _500Handler(ctx Context) {
	if e, ok := recover().(error); ok {
		ctx.write(fmt.Sprintf("500!\n%v", e))
	}
}

func (s *Server) Serve() {
	defer func() { fmt.Println("Shutting down...") }()

	addr := fmt.Sprint(s.ip, ":", s.port)
	listener, _ := net.Listen("tcp", addr)

	fmt.Printf("listening %s:%d\n", s.ip, s.port)

	http.Serve(listener, s)
}

func New(ip string, port int) Server {
	s := new(Server)
	s.ip = ip
	s.port = port

	s.initRoutes()

	return *s
}
