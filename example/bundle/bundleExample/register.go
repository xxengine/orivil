package bundleExample

import (
	"gopkg.in/orivil/router.v0"
	"gopkg.in/orivil/service.v0"
	"gopkg.in/orivil/middle.v0"
	"gopkg.in/orivil/orivil.v2"
)

type Register struct {}

// register bundle controllers
func(*Register) RegRoute(c *router.Container) {

	c.Add("{get}/", func() interface{} {

		return new(Controller)
	})
}

// register global services
func(*Register) RegService(c *service.Container) {}

// register global middleware
func(*Register) RegMiddle(c *middle.Container) {}

// configure global middleware
func(*Register) CfgMiddle(bag *middle.Bag) {}

// boot services after all services registered
func(*Register) Boot(s *orivil.Server) {}

// close services when server got terminate signal
func (*Register) Close() {}
