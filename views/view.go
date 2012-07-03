package views

import (
    "bitbucket.org/juztin/dingo"
)

var (
    viewCol = make(map[string]View)
)

type View interface {
    Name() string
    Associate(names ...string) error
    Extends(name string) error
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
        v.Execute(ctx, data)
    }

    // TODO - issue a 404
}


