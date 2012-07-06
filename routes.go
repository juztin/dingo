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

func NewSRoute(path string, handler Handler) Route {
	rt := new(route)
	rt.path = path
	rt.handler = handler

	return *rt
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

/*---------regexp2----------*/
type rRoute struct {
	path    string
	expr    *regexp.Regexp
	handler reflect.Value
}

func NewRRoute(re string, handler interface{}) Route {
	r := new(rRoute)
	r.path = re
	r.expr = regexp.MustCompile(re)

	if fn, ok := handler.(reflect.Value); ok {
		r.handler = fn
	} else {
		r.handler = reflect.ValueOf(handler)
	}

	return *r
}

func (r rRoute) Path() string {
	return r.path
}

func (r rRoute) Matches(url string) bool {
	return r.expr.MatchString(url)
}

func (r rRoute) Execute(ctx Context) {
	// TODO it would be nice if we could detect numbers and cast them as such prior to invoking the func
	args := []reflect.Value{reflect.ValueOf(ctx)}
	matches := r.expr.FindStringSubmatch(ctx.URL.Path)
	for _, a := range matches[1:] {
		args = append(args, reflect.ValueOf(a))
	}
	r.handler.Call(args)
}
