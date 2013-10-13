package views

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"text/template"

	"bitbucket.org/juztin/dingo"
)

var (
	viewCol = make(map[string]View)
)

type View interface {
	Name() string
	Associate(names ...string) error
	Associations() []View
	Extends(name string) error
	Extensions() []View
	Data(ctx dingo.Context) []byte
	Reload(ctx dingo.Context) error
	Save(ctx dingo.Context, data []byte) error
	Execute(ctx dingo.Context, data interface{}) error
}

func Add(key string, v View) {
	viewCol[key] = v
}
func Get(key string) View {
	if v, ok := viewCol[key]; ok {
		return v
	}
	return nil
}
func Execute(ctx dingo.Context, key string, data interface{}) {
	if v, ok := viewCol[key]; !ok {
		ctx.HttpError(404)
	} else if err := v.Execute(ctx, data); err != nil {
		// TODO log this somewhere
		log.Println("dingo: template execution error, ", err)
		/* This will cause a warning to be logged from `net/http/server.go`.
		 * The headers have, most likely, been written to the stream. The error is
		 * occuring midway through template processing, which is writing to the response stream.
		 * Server.go logs this; if we don't call the error handler below, then the stream is cut-off
		 * with no other warning to the client, with this they at-least get the 500 template.
		 */
		ctx.HttpError(500)
	}
}

type CoreView struct {
	IsStale              bool
	ViewName             string
	Associated, Extended []string
}

func (v *CoreView) Name() string {
	return v.ViewName
}
func (v *CoreView) Associate(names ...string) error {
	for _, n := range names {
		if view := Get(n); view != nil {
			v.Associated = append(v.Associated, view.Name())
		}
	}

	// TODO return an error
	return nil
}
func (v *CoreView) Associations() (views []View) {
	for _, n := range v.Associated {
		views = append(views, Get(n))
	}
	return
}
func (v *CoreView) Extends(name string) error {
	if view := Get(name); view != nil {
		v.IsStale = true
		v.Extended = append(v.Extended, view.Name())
		view.Associate(v.Name())
	}

	// TODO return an error
	return nil
}
func (v *CoreView) Extensions() (views []View) {
	for _, n := range v.Extended {
		views = append(views, Get(n))
	}
	return
}

/*----------------------------Common Templ Helpers----------------------------*/
func equals(x, y interface{}) bool {
	return x == y
}
func empty(o interface{}) bool {
	switch t := reflect.ValueOf(o); t.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return t.Len() == 0
	case reflect.String:
		return t.String() == ""
	}
	return true
}

var commonFuncs = template.FuncMap{
	"equals": equals,
	"join":   strings.Join,
	"empty":  empty,
}

func NewTmpl(name string) *template.Template {
	return template.New(name).Funcs(commonFuncs)
}
func AddTmplFunc(name string, fn interface{}) {
	commonFuncs[name] = fn
}

/*------------------Base Template, base on `text/template`--------------------*/
type TemplateData func(ctx dingo.Context, name string) (*template.Template, []byte, error)
type TemplateView struct {
	CoreView
	Tmpl     *template.Template
	TmplData TemplateData
	Bytes    []byte
}

func (v *TemplateView) Init(name string, dataFunc TemplateData) {
	v.ViewName = name
	v.Tmpl, _ = NewTmpl(name).Parse(EmptyTmpl)
	v.TmplData = dataFunc
	v.IsStale = true
}

func (v *TemplateView) Data(ctx dingo.Context) []byte {
	if v.IsStale {
		t, b, e := v.TmplData(ctx, v.ViewName)
		if e != nil {
			log.Println(e.Error())
			return []byte(e.Error())
		}
		v.Tmpl, v.Bytes = t, b
	}
	return v.Bytes
}
func reload(ctx dingo.Context, t *template.Template, v []View) error {
	for _, view := range v {
		if err := reload(ctx, t, view.Extensions()); err != nil {
			return err
		} else if t, err = t.Parse(string(view.Data(ctx))); err != nil {
			log.Printf("Failed to parse extension: (%s)\n%s\n", view.Name(), err)
		}
	}
	return nil
}
func (v *TemplateView) Reload(ctx dingo.Context) error {
	t, b, e := v.TmplData(ctx, v.ViewName)
	if e != nil {
		log.Println(e)
		return e
	}

	var view View
	// reload/re-parse all extensions
	reload(ctx, t, v.Extensions())
	v.Tmpl, v.Bytes = t, b

	// notify all associated templates
	for _, n := range v.Associated {
		if view = Get(n); view == nil {
			return errors.New(fmt.Sprintf("View doesn't exist: %s\n", n))
		}
		view.Reload(ctx)
	}
	v.IsStale = false

	return nil
}
func (v *TemplateView) Execute(ctx dingo.Context, data interface{}) error {
	if v.IsStale {
		v.Reload(ctx)
	}

	if v.Tmpl == nil {
		return errors.New("Template is `nil`")
	}

	return v.Tmpl.Execute(ctx.Response, data)
}
