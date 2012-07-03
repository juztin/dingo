package views

import (
    "fmt"
    "io/ioutil"
    "path/filepath"
    "text/template"

    "bitbucket.org/juztin/dingo"
)

var (
    TmplPath = "./templates"
    viewCol = make(map[string]dingo.View)
)

func Add(key string, v dingo.View) {
    viewCol[key] = v
}
func Get(key string) dingo.View {
    if v, ok := viewCol[key]; ok {
        return v
    }
    return nil
}
func Execute(ctx dingo.Context, key string, data interface{}) {
    if v, ok := viewCol[key]; ok {
        v.Execute(ctx, data)
    }

    // TODO - issue a 404
}

type view struct {
    tmpl *template.Template
    bytes []byte
    isLoaded bool
    loc string
    associated []dingo.View
    extensions []dingo.View
}

func parseFile(viewLoc string) (*template.Template, []byte, error) {
	b, err := ioutil.ReadFile(viewLoc)
	if err != nil {
		return nil, nil, err
	}

    p := filepath.Join(TmplPath, viewLoc)
    t := template.New(p)
	if _, err = t.Parse(string(b)); err != nil {
		return nil, nil, err
	}

	return t, b, nil
}

func New(location string) dingo.View {
    p := filepath.Join(TmplPath, location)
    v := new(view)
    v.loc = location
    v.tmpl, _ = template.New(p).Parse(EmptyTmpl)
    Add(location, v)

    return v
}

func (v *view) Name() string {
    return v.loc
}
func (v *view) Associate(names ...string) error {
    for _, n := range names {
        if view, ok := viewCol[n]; ok {
            v.associated = append(v.associated, view)
        }
    }

    // TODO return an error
    return nil
}
func (v *view) Extends(name string) error {
    /*var t *template.Template
    if t, err = v.tmpl.Clone(); err != nil {
        return err
    } else if t, err = t.Parse(view.Data()); err != nil {
        return err
    }
    v.tmpl = t*/

    if view, ok := viewCol[name]; ok {
        v.isLoaded = false
        v.extensions = append(v.extensions, view)
        view.Associate(v.Name())
    }

    // TODO return an error
    return nil
}
func (v *view) Data(ctx dingo.Context) string {
    if !v.isLoaded {
        p := filepath.Join(TmplPath, v.loc)
        t, b, e := parseFile(p)
        if e != nil {
            fmt.Println(e.Error())
            return e.Error()
        }
        v.tmpl, v.bytes = t, b
    }
    return string(v.bytes)
}
func (v *view) Reload(ctx dingo.Context) error {
    p := filepath.Join(TmplPath, v.loc)
    t, b, e := parseFile(p)
    if e != nil {
        fmt.Println(e.Error())
        return e
    }

    // reload/re-parse all extensions
    for _, v := range v.extensions {
        if t, e = t.Parse(v.Data(ctx)); e != nil {
            fmt.Println(e.Error())
            return e
        }
    }

    // notify all associated templates
    for _, v := range v.associated {
        v.Reload(ctx)
    }

    v.tmpl, v.bytes = t, b
    v.isLoaded = true

    return nil
}
func (v *view) Save(ctx dingo.Context, data []byte) error {
    t := template.New("")
    if _, err := t.Parse(string(data)); err != nil {
        return err
    }

    p := filepath.Join(TmplPath, v.loc)
    if err := ioutil.WriteFile(p, data, 0600); err != nil {
        return err
    }

    v.Reload(ctx)
    return nil
}
func (v *view) Execute(ctx dingo.Context, data interface{}) error {
    if !v.isLoaded {
        v.Reload(ctx)
    }

    return v.tmpl.Execute(ctx.Writer, data)
}
