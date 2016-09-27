package debug

import (
	"gopkg.in/orivil/router.v0"
	"gopkg.in/orivil/service.v0"
	"gopkg.in/orivil/middle.v0"
	"gopkg.in/orivil/orivil.v2"
)

type Register struct {}

func(*Register) RegRoute(c *router.Container) {

	c.Add("{get}/", func() interface{} {

		return new(Controller)
	})
}

func(*Register) RegService(c *service.Container) {}

func(*Register) RegMiddle(c *middle.Container) {

	c.Add(MidDebug, func(c *service.Container)interface{}{

		return &ViewComponent{}
	}, -32768)
}

func (*Register) CfgMiddle(bag *middle.Bag) {

	bag.Set(MidDebug).AllBundles().ExceptController("Controller")
}

func(*Register) Boot(s *orivil.Server) {}

func(*Register) Close() {}