package dingo

import (
	"regexp"
)

/*----------static----------*/
type route struct {
	path    string
	handler Handler
}

func NewRoute(path string, handler Handler) {
	rt := new(route)
	rt.path = path
	rt.handler = handler
}

func (r route) Path() string {
	return r.path
}
func (r route) Matches(path string) bool {
	return r.path == path
}

func (r route) Execute(ctx Context) {
	r.handler(ctx)
}

/*----------regexp----------*/
type reRoute struct {
	path    string
	expr    *regexp.Regexp
	handler Handler
}

func NewReRoute(re string, handler Handler) Route {
	rt := new(reRoute)
	rt.path = re
	rt.expr = regexp.MustCompile(re)
	rt.handler = handler

	return *rt
}

func (r reRoute) data(url string) map[string]string {
	data := make(map[string]string)
	matches := r.expr.FindAllStringSubmatch(url, -1)

	for i, n := range r.expr.SubexpNames() {
		if i == 0 {
			continue
		}
		data[n] = matches[0][i]
	}

	return data
}

func (r reRoute) Path() string {
	return r.path
}
func (r reRoute) Matches(url string) bool {
	return r.expr.MatchString(url)
}

func (r reRoute) Execute(ctx Context) {
	ctx.RouteData = r.data(ctx.URL.Path)
	r.handler(ctx)
}
