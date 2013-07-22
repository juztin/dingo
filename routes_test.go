package dingo

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"testing"
)

/*-------------------------Helpers-------------------------*/
func wc(ctx Context, s string) {
	ctx.Response.Write([]byte(s))
}
func dummyHandler(ctx Context)                   {}
func dummyHandler_0(ctx Context)                 { wc(ctx, ctx.Request.URL.Path) }
func dummyHandler_1(ctx Context, a string)       { wc(ctx, fmt.Sprintf("/%s/", a)) }
func dummyHandler_2(ctx Context, a, b string)    { wc(ctx, fmt.Sprintf("/%s/%s/", a, b)) }
func dummyHandler_3(ctx Context, a, b, c string) { wc(ctx, fmt.Sprintf("/%s/%s/%s/", a, b, c)) }

type rePath struct {
	ReExp, Path string
}

type sPath struct {
	RawPath, Path string
}

func sroutes(m map[sPath]Handler) map[string]Route {
	r := make(map[string]Route)
	for k, v := range m {
		r[k.RawPath] = NewSRoute(k.Path, v)
	}
	return r
}

func reroutes(m map[rePath]Handler) map[string]Route {
	r := make(map[string]Route)
	for k, v := range m {
		r[k.Path] = NewReRoute(k.ReExp, v)
	}
	return r
}

func rroutes(m map[rePath]interface{}) map[string]Route {
	r := make(map[string]Route)
	for k, v := range m {
		r[k.Path] = NewRRoute(k.ReExp, v)
	}
	return r
}

func dummyRequest(path string) *http.Request {
	req := &http.Request{Method: "GET"}
	req.URL, _ = url.Parse("http://www.juzt.in" + path)
	return req
}

type dummyWriter struct {
	buf *bytes.Buffer
}

func (w dummyWriter) Header() http.Header {
	return nil
}
func (w dummyWriter) Write(data []byte) (int, error) {
	return w.buf.Write(data)
}
func (w dummyWriter) WriteHeader(status int) {}
func (w dummyWriter) String() string {
	return w.buf.String()
}

/*-------------------------Data-------------------------*/
var srMatchCol = map[sPath]Handler{
	sPath{"/", "/"}:                                 dummyHandler_0,
	sPath{"/blog/", "/blog/"}:                       dummyHandler_0,
	sPath{"/a/b/c/", "/a/b/c/"}:                     dummyHandler_0,
	sPath{"/some-thing-cool/", "/some-thing-cool/"}: dummyHandler_0,
}
var srNoMatchCol = map[sPath]Handler{
	sPath{"/no/", "/"}:                             dummyHandler_0,
	sPath{"/blog", "/blog/"}:                       dummyHandler_0,
	sPath{"/a/b/c", "/a/b/c/"}:                     dummyHandler_0,
	sPath{"/some-thing-cool", "/some-thing-cool/"}: dummyHandler_0,
}

var reMatchCol = map[rePath]Handler{
	rePath{"^/$", "/"}:                                 dummyHandler_0,
	rePath{"^/blog/$", "/blog/"}:                       dummyHandler_0,
	rePath{"^/a/b/c/$", "/a/b/c/"}:                     dummyHandler_0,
	rePath{"^/some-thing-cool/$", "/some-thing-cool/"}: dummyHandler_0,
}
var reNoMatchCol = map[rePath]Handler{
	rePath{"^/$", ""}:                                 dummyHandler_0,
	rePath{"^/blog$", "/blog/"}:                       dummyHandler_0,
	rePath{"^/a/b/c$", "/a/b/c/"}:                     dummyHandler_0,
	rePath{"^/some-thing-cool$", "/some-thing-cool/"}: dummyHandler_0,
}

var rrMatchCol = map[rePath]interface{}{
	rePath{"^/$", "/"}:                                 dummyHandler_0,
	rePath{"^/blog/$", "/blog/"}:                       dummyHandler_0,
	rePath{"^/a/b/c/$", "/a/b/c/"}:                     dummyHandler_0,
	rePath{"^/some-thing-cool/$", "/some-thing-cool/"}: dummyHandler_0,
}
var rrNoMatchCol = map[rePath]interface{}{
	rePath{"^/$", ""}:                                 dummyHandler_0,
	rePath{"^/blog$", "/blog/"}:                       dummyHandler_0,
	rePath{"^/a/b/c$", "/a/b/c/"}:                     dummyHandler_0,
	rePath{"^/some-thing-cool$", "/some-thing-cool/"}: dummyHandler_0,
}
var rrArgCol = map[rePath]interface{}{
	rePath{"^/$", "/"}:                      dummyHandler_0,
	rePath{"^/(.*)/$", "/a/"}:               dummyHandler_1,
	rePath{"^/(.*)/(.*)/$", "/a/b/"}:        dummyHandler_2,
	rePath{"^/(.*)/(.*)/(.*)/$", "/a/b/c/"}: dummyHandler_3,
}

/*-------------------------Tests-------------------------*/
func TestSRouteMatches(t *testing.T) {
	sr := sroutes(srMatchCol)
	for path, r := range sr {
		if !r.Matches(path) {
			t.Errorf("Route doesn't match it's path: (%s) != (%s)", r.Path(), path)
		}
	}
}

func TestSRouteNoMatches(t *testing.T) {
	sr := sroutes(srNoMatchCol)
	for path, r := range sr {
		if r.Matches(path) {
			t.Errorf("Route doesn't match it's path: (%s) != (%s)", r.Path(), path)
		}
	}
}

func TestSRouteExecute(t *testing.T) {
	sr := sroutes(srMatchCol)
	for path, r := range sr {
		ctx := *new(Context)
		ctx.Response = dummyWriter{new(bytes.Buffer)}
		ctx.Request = dummyRequest(path)
		r.Execute(ctx)

		if ctx.Response.(dummyWriter).String() != r.Path() {
			t.Errorf("Handler recieved non-matching args: %v != %s", ctx.Response, r.Path())
		}
	}
}

func TestReRouteMatches(t *testing.T) {
	re := reroutes(reMatchCol)
	for path, r := range re {
		if !r.Matches(path) {
			t.Errorf("Route doesn't match it's path: (%s) != (%s)", r.Path(), path)
		}
	}
}

func TestReRouteNoMatches(t *testing.T) {
	re := reroutes(reNoMatchCol)
	for path, r := range re {
		if r.Matches(path) {
			t.Errorf("Route matches invalid path: (%s) == (%s)", r.Path(), path)
		}
	}
}

func TestRRouteMatches(t *testing.T) {
	rr := rroutes(rrMatchCol)
	for path, r := range rr {
		if !r.Matches(path) {
			t.Errorf("Route doesn't match it's path: (%s) != (%s)", r.Path(), path)
		}
	}
}

func TestRRouteNoMatches(t *testing.T) {
	rr := rroutes(rrNoMatchCol)
	for path, r := range rr {
		if r.Matches(path) {
			t.Errorf("Route matches invalid path: (%s) == (%s)", r.Path(), path)
		}
	}
}

func TestRRRouteExecute(t *testing.T) {
	rr := rroutes(rrArgCol)
	for path, r := range rr {
		ctx := *new(Context)
		ctx.Response = dummyWriter{new(bytes.Buffer)}
		ctx.Request = dummyRequest(path)
		r.Execute(ctx)

		if ctx.Response.(dummyWriter).String() != path {
			t.Errorf("Handler recieved non-matching args: %v != %s", ctx.Response, path)
		}
	}
}
