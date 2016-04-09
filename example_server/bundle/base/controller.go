package base

import (
	"gopkg.in/orivil/orivil.v0"
)

type Controller struct {
	*orivil.App
}

// @route {get}/
func (this *Controller) Index() {

	this.WriteString("<h1>hello orivil!</h1>")
}
