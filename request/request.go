package request

import "net/http"

func is(r *http.Request, s string) bool {
	ct := http.CanonicalHeaderKey(r.Header.Get("Content-Type"))
	l := len(s)
	return len(ct) >= l && ct[:l] == s
}

func IsApplicationJson(r *http.Request) bool {
	return is(r, "Application/json")
}

func IsTextHTML(r *http.Request) bool {
	return is(r, "Text/html")
}
