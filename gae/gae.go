package gae

import (
	"net/http"

	"bitbucket.org/juztin/dingo"
)

type handler func(w http.ResponseWriter, r *http.Request)
type GAEServer struct {
	dingo.Server
	fn handler
}

func gaeHandler(s http.Handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		s.ServeHTTP(w, r)
	}
}
func Server() GAEServer {
	d := dingo.New("", -1)
	return GAEServer{d, gaeHandler(&d)}
}
func (s *GAEServer) Serve() {
	http.HandleFunc("/", s.fn)
}