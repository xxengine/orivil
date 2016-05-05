// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orivil

import (
	"gopkg.in/orivil/middle.v0"
	"gopkg.in/orivil/router.v0"
	"gopkg.in/orivil/service.v0"
	"gopkg.in/orivil/session.v0"
)

const (
	SvcMemorySession    = "session.MemorySession"
	SvcPermanentSession = "session.PermanentSession"
	SvcSessionContainer = "orivil.SessionContainer"
	SessionContainerKey = "orivil.SessionContainerKey"
)

type BaseRegister struct{}

func (*BaseRegister) RegisterService(c *service.Container) {

	// register memory session as service
	c.Add(SvcMemorySession, func(c *service.Container) interface{} {
		app := c.Get(SvcApp).(*App)
		return session.NewMemorySession(app.Response, app.Request)
	})

	// register permanent session as service
	c.Add(SvcPermanentSession, func(c *service.Container) interface{} {
		app := c.Get(SvcApp).(*App)
		return session.NewPermanentSession(app.Response, app.Request)
	})

	// register session container as service
	c.Add(SvcSessionContainer, func(c *service.Container) interface{} {
		session := c.Get(SvcMemorySession).(*session.Session)

		// get session container form memory session
		sessionContainer := session.GetData(SessionContainerKey)

		var private *service.Container
		if sessionContainer == nil {
			private = service.NewPrivateContainer(c.Public)
			session.SetData(SessionContainerKey, private)
		} else {
			private = sessionContainer.(*service.Container)
		}
		return private
	})
}

func (*BaseRegister) RegisterRoute(c *router.Container) {}

func (*BaseRegister) RegisterMiddle(c *middle.Container) {}

func (*BaseRegister) Boot(c *service.Container) {}
