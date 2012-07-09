package gae

import (
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
	k := datastore.NewKey(c, "Template", key, 0, nil)
	tb := new(TemplateBytes)

	if err := datastore.Get(c, k, tb); err != nil {
		if err != datastore.ErrNoSuchEntity {
			return []byte(emptyTempl), nil
		}
		fmt.Printf("dingo [TEMPLATE_BIGTABLE_ERR] / {%v} - %v\n", key, err)
		return nil, err
	}

	return tb.Bytes, nil
}
func getTemplate(ctx dingo.Context, key string) (t *template.Template, b []byte, err error) {
	c := appengine.NewContext(ctx.Request)
	if b, err = getTemplateBytes(c, key); err == nil {
		t = template.New(key)
		t, err = t.Parse(string(b))
	}

	return
}

type gae struct {
	views.TemplateView
}

func New(key string) views.View {
	g := new(gae)
	g.Init(key, getTemplate)
	views.Add(key, g)

	return g
}
func (g *gae) Save(ctx dingo.Context, data []byte) error {
	t := template.New("")
	if _, err := t.Parse(string(data)); err != nil {
		return err
	}

	b := &TemplateBytes{data}
	c := appengine.NewContext(ctx.Request)
	k := datastore.NewKey(c, "Template", g.ViewName, 0, nil)

	if _, err := datastore.Put(c, k, b); err != nil {
		fmt.Printf("dingo [TEMPLATE_BIGTABLE_SAVE_ERR] / {%v} - %v\n", g.ViewName, err)
	}

	g.Reload(ctx)
	return nil
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
	return GAEServer{d, gaeHandler(&d)}
}
func (s *GAEServer) Serve() {
	http.HandleFunc("/", s.fn)
}
