package bundleExample

import (
	"gopkg.in/orivil/orivil.v2"
)

type Controller struct {

	*orivil.App
}

// @route {get}/
func (c *Controller) Index() {

	c.View().With("a1", "Orivil!")
}