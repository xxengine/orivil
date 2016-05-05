// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orivil

type redirect struct {
	url  string
	code int
}

type end struct {}

func Return() {
	panic(&end{})
}

func RedirectCode(url string, code int) {
	panic(&redirect{url: url, code: code})
}

func Redirect(url string) {
	panic(&redirect{url: url, code: 302})
}
