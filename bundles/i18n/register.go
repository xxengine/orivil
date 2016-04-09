package i18n

import (
	"github.com/orivil/event"
	"github.com/orivil/middle"
	"github.com/orivil/orivil"
	"github.com/orivil/router"
	"github.com/orivil/service"
	"path/filepath"
)

type Register struct {
}

func (*Register) RegisterRoute(c *router.Container) {
	c.Add("{get}/", func() interface{} { return new(Controller) })
}

func (*Register) SetMiddle(bag *middle.Bag) {

	bag.Set(MidViewFileReader).AllBundles().ExceptController("Controller")
}

func (*Register) RegisterService(c *service.Container) {}

func (*Register) RegisterMiddle(c *middle.Container) {

	c.Add(MidDataSender, DataSender)
	c.Add(MidViewFileReader, ViewDirReader, 10000)
}

func (*Register) Boot(c *service.Container) {}

func (*Register) AddServerListener(d *event.Dispatcher) {

	// auto generate I18n view files
	d.AddListener(new(Listener))
}
