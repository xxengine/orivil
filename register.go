// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orivil

import (
	"gopkg.in/orivil/middle.v0"
	"gopkg.in/orivil/router.v0"
	"gopkg.in/orivil/service.v0"
)

// Every bundle register should implement Register interface.
type Register interface {
	RegRoute(c *router.Container)

	RegService(c *service.Container)

	RegMiddle(c *middle.Container)

	CfgMiddle(bag *middle.Bag)

	Boot(s *Server)

	Close()
}

// MiddlewareConfigure provide for controllers.
type MiddlewareConfigure interface {

	CfgMiddle(bag *middle.Bag)
}