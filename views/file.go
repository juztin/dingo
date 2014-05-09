// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package views

import (
	"io/ioutil"
	"path/filepath"
	"text/template"

	"code.minty.io/dingo"
)

var (
	Path = "./templates"
)

// FileView reads a template from the file system.
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

// New returns a new FileView
func New(location string) View {
	v := new(FileView)
	v.Init(location, parseFile)
	Add(location, v)

	return v
}

// NewEditable returns a new editable FileView
func NewEditable(location string) View {
	return Editable(New(location))
}

// Save writes the new template data to the file system
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
