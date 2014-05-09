// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gae

import (
	"net/http"

	"code.minty.io/dingo"
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
