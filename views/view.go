package views

import (
	"errors"
	"fmt"
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
	Data(ctx dingo.Context) string
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
	if v, ok := viewCol[key]; ok {
		if err := v.Execute(ctx, data); err != nil {
			// TODO log
			fmt.Println(err)
		}
	}

	// TODO - issue a 404
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

type TemplateData func(ctx dingo.Context, name string) (*template.Template, []byte, error)
type TemplateView struct {
	CoreView
	Tmpl     *template.Template
	TmplData TemplateData
	Bytes    []byte
}

func (v *TemplateView) Init(name string, dataFunc TemplateData) {
	v.ViewName = name
	v.Tmpl, _ = template.New(name).Parse(EmptyTmpl)
	v.TmplData = dataFunc
	v.IsStale = true
}
func (v *TemplateView) Data(ctx dingo.Context) string {
	if v.IsStale {
		t, b, e := v.TmplData(ctx, v.ViewName)
		if e != nil {
			fmt.Println(e.Error())
			return e.Error()
		}
		v.Tmpl, v.Bytes = t, b
	}
	return string(v.Bytes)
}
func reload(ctx dingo.Context, t *template.Template, v []View) error {
	for _, view := range v {
		if err := reload(ctx, t, view.Extensions()); err != nil {
			return err
		} else if t, err = t.Parse(view.Data(ctx)); err != nil {
			fmt.Printf("Failed to parse extension: (%s)\n%s\n", view.Name(), err)
		}
	}
	return nil
}
func (v *TemplateView) Reload(ctx dingo.Context) error {
	t, b, e := v.TmplData(ctx, v.ViewName)
	if e != nil {
		fmt.Println(e)
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

	return v.Tmpl.Execute(ctx.Writer, data)
}
