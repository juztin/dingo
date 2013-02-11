package dingo

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	//"reflect"
	"runtime"
)

const (
	VERSION string = "0.1.5"
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
	Writer    http.ResponseWriter
	RouteData map[string]string
}

func newContext(request *http.Request, writer http.ResponseWriter) Context {
	c := new(Context)
	c.Request, c.Writer = request, writer
	return *c
}

func (c *Context) write(content string) {
	//b := []byte(content)
	//c.Writer.Write(b)
	c.Writer.Write([]byte(content))
}

func (c *Context) Redirect(path string) {
	http.Redirect(c.Writer, c.Request, path, http.StatusTemporaryRedirect)
}

func (c *Context) RedirectPerm(path string) {
	w := c.Writer
	w.Header().Set("Location", path)
	w.WriteHeader(http.StatusMovedPermanently)
}

func (c *Context) HttpError(status int, msg ...[]byte) {
	if ErrorHandler != nil {
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
	w := c.Writer
	w.WriteHeader(status)
	if msg == nil {
		m := []byte(http.StatusText(status))
		w.Write(m)
	} else {
		for _, m := range msg {
			w.Write(m)
		}
	}
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

/*----------------------------------Router------------------------------------*/
type newRoute func(path string, h Handler) Route

func iroute(route newRoute) NewIRoute {
	var fn NewIRoute
	fn = func(path string, h interface{}) Route {
		/*if h, ok := h.(Handler); ok {
			return route(path, h)
		}
		if h, ok := h.(func(Context)); ok {
			return route(path, Handler(h))
		}
		t := reflect.TypeOf(h)
		panic(fmt.Sprintf("Handler type: %v is an invalid handler", t))*/

		//switch t := h.(type) {
		switch h.(type) {
		case Handler:
			return route(path, h.(Handler))
		case func(Context):
			return route(path, Handler(h.(func(Context))))
		}

		//t := reflect.TypeOf(h)
		//panic(fmt.Sprintf("Handler type: %v is an invalid handler", t))
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

/* http.Handler */
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w)
	defer _500Handler(ctx)

	if path, ok := isCanonical(r.URL.Path); !ok {
		ctx.RedirectPerm(path)
		return
	}

	if routes, ok := s.routes[r.Method]; ok {
		if r, ok := routes.Route(r.URL.Path); ok {
			r.Execute(ctx)
			return
		}
	}

	ctx.HttpError(404, nil)
}

/*--------------*/
func _500Handler(ctx Context) {
	//if err, ok := recover().(error); ok {
	if err := recover(); err != nil {
		// TODO write the 500 message, or stack, depending on some settings
		// if !emitError
		//ctx.Writer.WriteHeader(http.StatusInternalServerError)
		//ctx.write(http.StatusText(500))
		ctx.HttpError(500, nil)

		// else
		// hmm.. `i` doesn't get incremented on the call to `runtime.Caller(i)`
		//i := 1
		//for _, f, l, ok := runtime.Caller(i); ok; i++ {
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

func (s *Server) Serve() {
	defer func() {
		fmt.Println("Shutting down...")
		s.listener.Close()
	}()

	http.Serve(s.listener, s)
}

func New(l net.Listener) Server {
	s := new(Server)
	s.initRoutes()
	s.listener = l

	return *s
}

func HttpHandler(ip string, port int) (net.Listener, error) {
	addr := fmt.Sprint(ip, ":", port)
	return net.Listen("tcp", addr)
}

func TLSHandler(ip string, port int, certFile, keyFile string) (net.Listener, error) {
	// Most of this is copied from Go source `net/http - server.go`
	addr := fmt.Sprint(ip, ":", port)
	config := &tls.Config{NextProtos: []string{"http/1.1"}}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return conn, err
	}

	return tls.NewListener(conn, config), nil
}

func SOCKHandler(sockFile string, mode os.FileMode) (net.Listener, error) {
	// delete stale sock
	// TODO check errors other than file doesn't exist
	os.Remove(sockFile)

	// create UNIX sock
	sock, err := net.ResolveUnixAddr("unix", sockFile)
	if err != nil {
		return nil, err
	}
	if l, err := net.ListenUnix("unix", sock); err == nil {
		err = os.Chmod(sockFile, mode)
		return l, err
	}
	return nil, err
}
