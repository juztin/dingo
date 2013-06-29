package gae

import (
	"net/http"

	"bitbucket.org/juztin/dingo"
)

type GAEServer struct {
	dingo.Server
}

func Server() GAEServer {
	return GAEServer{dingo.New(nil)}
}

func (s *GAEServer) Serve() {
	http.HandleFunc("/", s.ServeHTTP)
}
