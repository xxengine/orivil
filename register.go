package orivil

import (
	"github.com/orivil/event"
	"github.com/orivil/middle"
	"github.com/orivil/router"
	"github.com/orivil/service"
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

// example:
//
// import (
//     "orivil/router"
//     "orivil/middle"
//     "orivil/service"
// )
//
// type Register struct {}

// func(*Register) RegisterRoute(c *router.Container) {}
//
// func(*Register) RegisterService(c *service.Container) {}
//
// func(*Register) RegisterMiddle(c *middle.Container) {}
//
// func(*Register) Boot(c *service.Container) {}
//
// func(*Register) SetMiddle(bag *middle.Bag) {}
