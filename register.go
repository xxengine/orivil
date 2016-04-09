package orivil

import (
	"github.com/orivil/event.v0"
	"github.com/orivil/middle.v0"
	"github.com/orivil/router.v0"
	"github.com/orivil/service.v0"
)

// every bundle register should implement Register interface
type Register interface {
	RegisterRoute(c *router.Container)

	RegisterService(c *service.Container)

	RegisterMiddle(c *middle.Container)

	Boot(c *service.Container)
}

type MiddlewareConfigure interface {
	SetMiddle(bag *middle.Bag)
}

type ServerEventListener interface {
	AddServerListener(d *event.Dispatcher)
}
