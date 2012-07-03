package gae

import (
    "errors"
    "fmt"
    "net/http"
    "text/template"

    "appengine"
    "appengine/datastore"
    "bitbucket.org/juztin/dingo"
    "bitbucket.org/juztin/dingo/views"
)

var emptyTempl string = "<!doctype html><html><head><title>New Page</title></head><body>Page Hasn't Been Created</body></html>"

type TemplateBytes struct {
    Bytes []byte
}

func getTemplateBytes(c appengine.Context, key string) ([]byte, error) {
	k  := datastore.NewKey(c, "Template", key, 0, nil)
	tb := new(TemplateBytes)

	if err := datastore.Get(c, k, tb); err != nil {
        if err != datastore.ErrNoSuchEntity {
            return []byte(emptyTempl), nil
		}
        fmt.Println(fmt.Sprintf("dingo [TEMPLATE_BIGTABLE_ERR] / {%v} - %v", key, err))
        return nil, err
    }

    return tb.Bytes, nil
}
func getTemplate(c appengine.Context, key string) (t *template.Template, b []byte, err error) {
    if b, err = getTemplateBytes(c, key); err == nil {
        t = template.New(key)
        t, err = t.Parse(string(b))
    }

    return
}

type gae struct {
    isStale bool
    tmpl *template.Template
    bytes []byte
    key string
    associated, extensions []string
}

func New(key string) views.View {
    g := new(gae)
    g.key = key
    g.tmpl, _ = template.New(key).Parse(views.EmptyTmpl)
    g.isStale = true
    views.Add(key, g)

    return g
}

func (g *gae) Name() string {
    return g.key
}
func (g *gae) Associate(names ...string) error {
    for _, n := range names {
        if view := views.Get(n); view != nil {
            g.associated = append(g.associated, view.Name())
        }
    }

    // TODO return an error
    return nil
}
func (g *gae) Extends(name string) error {
    if view := views.Get(name); view != nil {
        g.isStale = true
        g.extensions = append(g.extensions, view.Name())
        view.Associate(g.Name())
    }

    // TODO return an error
    return nil
}
func (g *gae) Data(ctx dingo.Context) string {
    if g.isStale {
        c := appengine.NewContext(ctx.Request)
        t, b, e := getTemplate(c, g.key)
        if e != nil {
            return e.Error()
        }
        g.tmpl, g.bytes = t, b
    }
    return string(g.bytes)
}
func (g *gae) Reload(ctx dingo.Context) error {
    c := appengine.NewContext(ctx.Request)
    t, b, e := getTemplate(c, g.key)
    if e != nil {
        fmt.Println(e.Error())
        return e
    }

    var v views.View
    // reload/re-parse all extensions
    for _, n := range g.extensions {
        if v = views.Get(n); v == nil {
            return errors.New(fmt.Sprintf("View doesn't exist: %s\n", n))
        }
        if t, e = t.Parse(v.Data(ctx)); e != nil {
            fmt.Println(e.Error())
            return e
        }
    }
    g.tmpl, g.bytes = t, b

    // notify all associated templates
    for _, n := range g.associated {
        if v := views.Get(n); v == nil {
            return errors.New(fmt.Sprintf("View doesn't exist: %s\n", n))
        }
        v.Reload(ctx)
    }
    g.isStale = false

    return nil
}
func (g *gae) Save(ctx dingo.Context, data []byte) error {
	t := template.New("")
	if _, err := t.Parse(string(data)); err != nil {
		return err
	}

	b := &TemplateBytes{data}
	c := appengine.NewContext(ctx.Request)
	k := datastore.NewKey(c, "Template", g.key, 0, nil)

	if _, err := datastore.Put(c, k, b); err != nil {
		fmt.Println(fmt.Sprintf("dingo [TEMPLATE_BIGTABLE_SAVE_ERR] / {%v} - %v", g.key, err))
	}

	g.Reload(ctx)
	return nil
}
func (g *gae) Execute(ctx dingo.Context, data interface{}) error {
    if g.isStale {
		g.Reload(ctx)
	}

	if g.tmpl == nil {
		return errors.New("Template is `nil`")
	}

	return g.tmpl.Execute(ctx.Writer, data)
}



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
    return GAEServer{ d, gaeHandler(&d) }
}
func (s *GAEServer) Serve() {
    http.HandleFunc("/", s.fn)
}
