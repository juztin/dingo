// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package request

import "net/http"

func is(r *http.Request, s string) bool {
	h := r.Header.Get("Content-Type")
	if h == "" {
		h = r.Header.Get("Accept")
	}
	ct := http.CanonicalHeaderKey(h)
	l := len(s)
	return len(ct) >= l && ct[:l] == s
}

func IsApplicationJson(r *http.Request) bool {
	return is(r, "Application/json")
}

func IsTextHTML(r *http.Request) bool {
	return is(r, "Text/html")
}
