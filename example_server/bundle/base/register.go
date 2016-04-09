package base

import (
	"github.com/orivil/router"
	"github.com/orivil/service"
	"github.com/orivil/middle"
	"github.com/orivil/event"
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
