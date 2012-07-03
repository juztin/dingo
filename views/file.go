package views

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"bitbucket.org/juztin/dingo"
)

var (
	Path = "./templates"
)

type view struct {
	tmpl                   *template.Template
	bytes                  []byte
	isLoaded               bool
	loc                    string
	associated, extensions []string
}

func parseFile(viewLoc string) (*template.Template, []byte, error) {
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
	p := filepath.Join(Path, location)
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
		if view := Get(n); view != nil {
			v.associated = append(v.associated, view.Name())
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

	if view := Get(name); view != nil {
		v.isLoaded = false
		//v.extensions = append(v.extensions, view)
		v.extensions = append(v.extensions, view.Name())
		view.Associate(v.Name())
	}

	// TODO return an error
	return nil
}
func (v *view) Data(ctx dingo.Context) string {
	if !v.isLoaded {
		p := filepath.Join(Path, v.loc)
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
	p := filepath.Join(Path, v.loc)
	t, b, e := parseFile(p)
	if e != nil {
		fmt.Println(e.Error())
		return e
	}

	var x View
	// reload/re-parse all extensions
	for _, n := range v.extensions {
		if x = Get(n); x == nil {
			return errors.New(fmt.Sprintf("View doesn't exist: %s\n", n))
		}
		if t, e = t.Parse(x.Data(ctx)); e != nil {
			fmt.Println(e.Error())
			return e
		}
	}

	// update content
	v.tmpl, v.bytes = t, b
	v.isLoaded = true

	// notify all associated templates
	for _, n := range v.associated {
		if x = Get(n); x == nil {
			return errors.New(fmt.Sprintf("View doesn't exist: %s\n", n))
		}
		x.Reload(ctx)
	}

	return nil
}
func (v *view) Save(ctx dingo.Context, data []byte) error {
	t := template.New("")
	if _, err := t.Parse(string(data)); err != nil {
		return err
	}

	p := filepath.Join(Path, v.loc)
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
