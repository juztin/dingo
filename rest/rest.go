package rest

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"

	"minty.io/dingo"
)

type Handler func(ctx dingo.Context) (int, interface{})

type Wrapper func(fn Handler) dingo.Handler

type encoder func(o interface{}) ([]byte, error)

var MaxMsgSize int64 = 1 << 20

func pad(data []byte, callback string) []byte {
	l := len(callback)
	s := make([]byte, len(data)+l+2)
	copy(s[l+1:], data)
	copy(s[0:], []byte(callback))
	s[l] = '('
	s[len(s)-1] = ')'
	return s
}

func JSONHandler(fn Handler, ctx dingo.Context) {
	var (
		enc      encoder
		callback string
	)

	ct := ctx.Header.Get("Accept")
	ct, _, _ = mime.ParseMediaType(ct)
	switch ct {
	default:
		ct = "application/json"
		callback = ctx.URL.Query().Get("callback")
		if len(callback) > 0 {
			ct = "application/javascript"
		}
		enc = json.Marshal
	case "application/json":
		enc = json.Marshal
	case "application/xml":
		enc = xml.Marshal
	}

	ctx.Response.Header().Set("Content-Type", ct)
	status, o := fn(ctx)
	ctx.Response.WriteHeader(status)
	if o == nil && len(callback) < 1 {
		return
	}

	data, err := enc(o)
	if err != nil {
		ctx.HttpError(http.StatusInternalServerError)
		return
	}

	if len(callback) > 0 {
		data = pad(data, callback)
	}
	ctx.Response.Write(data)
}

func Wrap(fn Handler) dingo.Handler {
	return func(ctx dingo.Context) {
		JSONHandler(fn, ctx)
	}
}

func Data(ctx dingo.Context) ([]byte, error) {
	if ctx.Method != "POST" && ctx.Method != "PUT" {
		return nil, nil
	}

	defer ctx.Body.Close()
	reader := http.MaxBytesReader(ctx.Response, ctx.Body, MaxMsgSize)
	return ioutil.ReadAll(reader)
}

func JSONData(ctx dingo.Context, o interface{}) error {
	b, err := Data(ctx)
	if err == nil {
		err = json.Unmarshal(b, &o)
	}
	return err
}

func JSONDataMap(ctx dingo.Context) (map[string]interface{}, error) {
	var o interface{}
	if err := JSONData(ctx, &o); err != nil {
		return nil, err
	}
	j, ok := o.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert JSON to map[string]interface{}, %s", o)
	}
	return j, nil
}
