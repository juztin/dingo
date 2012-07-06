package views

import (
	"testing"

	"bitbucket.org/juztin/dingo"
)

type dummyView struct {
	name string
}

func (d *dummyView) Name() string {
	return d.name
}
func (d *dummyView) Associate(names ...string) error {
	return nil
}
func (d *dummyView) Extends(name string) error {
	return nil
}
func (d *dummyView) Data(ctx dingo.Context) string {
	return ""
}
func (d *dummyView) Reload(ctx dingo.Context) error {
	return nil
}
func (d *dummyView) Save(ctx dingo.Context, data []byte) error {
	return nil
}
func (d *dummyView) Execute(ctx dingo.Context, data interface{}) error {
	return nil
}
func newDummy(name string) View {
	d := new(dummyView)
	d.name = name
	return d
}

type AddTest struct {
	name string
	view View
}

var addData = []AddTest{
	{"", newDummy("")},
	{"a", newDummy("a")},
	{"1", newDummy("1")},
	{"./some/path/index.html", newDummy("./some/path/index.html")},
	{"./some/path/世界", newDummy("./some/path/世界")},
	{"Καλημέρα κόσμε; or こんにちは 世界", newDummy("Καλημέρα κόσμε; or こんにちは 世界")},
}

func TestAdd(t *testing.T) {
	for _, d := range addData {
		Add(d.name, d.view)
		if v, ok := viewCol[d.name]; !ok {
			t.Errorf("Failed to add(%s)", d.name)
		} else if v != d.view {
			t.Error("Unexpected view from collection")
		}
	}
}

func TestGet(t *testing.T) {
	for _, d := range addData {
		Add(d.name, d.view)
		if v := Get(d.name); v != d.view {
			t.Errorf("Failed to get(%s)")
		}
	}
}
