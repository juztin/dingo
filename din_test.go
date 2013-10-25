// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dingo

import (
	"testing"
)

var canonicalPaths = []string{
	"/",
	"/some/path/",
}
var nonCanonicalPaths = []string{
	"",
	"/some/path",
}

func TestIsCanonical(t *testing.T) {
	for _, path := range canonicalPaths {
		if cPath, ok := IsCanonical(path); !ok {
			t.Errorf("Path is canonical (%s)", path)
		} else if path != cPath {
			t.Errorf("Canonical paths are not the same (%s):(%s)", path, cPath)
		}
	}
}

func TestIsNotCanonical(t *testing.T) {
	for _, path := range nonCanonicalPaths {
		if cPath, ok := IsCanonical(path); ok {
			t.Errorf("Path is canonical (%s)", path)
		} else if path == cPath {
			t.Errorf("Non Canonical path is the same (%s):(%s)", path, cPath)
		} else if cPath != path+"/" {
			t.Errorf("Unexpected canonical path: (%s), from :(%s)", cPath, path)
		}
	}
}
