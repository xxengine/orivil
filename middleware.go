// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orivil

type RequestHandler interface {
	Handle(app *App)
}

type TerminateHandler interface {
	Terminate(app *App)
}
