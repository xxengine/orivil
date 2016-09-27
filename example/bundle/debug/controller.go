package debug

import (
	"gopkg.in/orivil/orivil.v2"
)

type Controller struct {

	*orivil.App
}

// @route {get}/debug/history
func (this *Controller) History() {

	this.JsonEncode(history)
}

// @route {get}/debug/sqlQuery
func (this *Controller) SQLs() {

	this.JsonEncode(SQLs)
}

// @route {get}/debug/mergedHtml
func (this *Controller) MergedHtml() {

	this.Response.Write(mergedHtml)
}