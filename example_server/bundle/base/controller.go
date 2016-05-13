package base

import (
	"gopkg.in/orivil/orivil.v1"
)

type Controller struct {

	*orivil.App
}

// @route {get}/
func (this *Controller) Index() {

	this.View().With("say", "Orivil!")
}

// @route {get}/set-session/::value
func (this *Controller) SetSession() {

	getSession := this.Session().Get("name")

	setSession := this.Params["value"]

	this.Session().Set("name", setSession)

	this.View("index")

	this.With("say", "Orivil!")
	this.With("getSession", getSession)
	this.With("setSession", setSession)
}