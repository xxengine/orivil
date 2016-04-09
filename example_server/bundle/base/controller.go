package base

import (
	"fmt"
	"github.com/orivil/orivil"
)

type Controller struct {
	*orivil.App
}

// @route {get}/
func (this *Controller) Index() {

	this.WriteString("<h1>hello orivil!</h1>")
}

// search user
//
// @route {get|post}/search/user/::country/::username
// @route {get|post}/search/user/::username
func (this *Controller) Search() {

	name := this.Params["username"]
	country := this.Params["country"]
	this.WriteString(fmt.Sprintf("country: %s", country))
	this.WriteString(fmt.Sprintf("name: %s", name))
}
