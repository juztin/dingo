package views

import (
	"io/ioutil"
	"path/filepath"
	"text/template"

	"bitbucket.org/juztin/dingo"
)

var (
	Path = "./templates"
)

type view struct {
	TemplateView
}

func parseFile(ctx dingo.Context, name string) (*template.Template, []byte, error) {
	viewLoc := filepath.Join(Path, name)
	b, err := ioutil.ReadFile(viewLoc)
	if err != nil {
		return nil, nil, err
	}

	p := filepath.Join(Path, viewLoc)
	t := template.New(p)
	if _, err = t.Parse(string(b)); err != nil {
		return nil, nil, err
	}

	return t, b, nil
}

func New(location string) View {
	v := new(view)
	v.Init(location, parseFile)
	Add(location, v)

	return v
}

func (v *view) Save(ctx dingo.Context, data []byte) error {
	t := template.New("")
	if _, err := t.Parse(string(data)); err != nil {
		return err
	}

	p := filepath.Join(Path, v.ViewName)
	if err := ioutil.WriteFile(p, data, 0600); err != nil {
		return err
	}

	v.Reload(ctx)
	return nil
}

