package base

import (
	"gopkg.in/orivil/router.v0"
	"gopkg.in/orivil/service.v0"
	"gopkg.in/orivil/middle.v0"
	"gopkg.in/orivil/event.v0"
)

type Register struct {}

func(*Register) RegisterRoute(c *router.Container) {

	c.Add("{get}/", func() interface{} {

		return new(Controller)
	})
}

func(*Register) RegisterService(c *service.Container) {}

func(*Register) RegisterMiddle(c *middle.Container) {}

func(*Register) Boot(c *service.Container) {}

func (*Register) SetMiddle(bag *middle.Bag) {}

func (*Register) AddServerListener(d *event.Dispatcher) {}
