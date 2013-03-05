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

type FileView struct {
	TemplateView
}

func parseFile(ctx dingo.Context, name string) (*template.Template, []byte, error) {
	viewLoc := filepath.Join(Path, name)
	b, err := ioutil.ReadFile(viewLoc)
	if err != nil {
		return nil, nil, err
	}

	p := filepath.Join(Path, viewLoc)
	t := NewTmpl(p)
	if _, err = t.Parse(string(b)); err != nil {
		return nil, nil, err
	}

	return t, b, nil
}

func New(location string) View {
	v := new(FileView)
	v.Init(location, parseFile)
	Add(location, v)

	return v
}

func NewEditable(location string) View {
	return Editable(New(location))
}

func (v *FileView) Save(ctx dingo.Context, data []byte) error {
	t := NewTmpl("")
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
