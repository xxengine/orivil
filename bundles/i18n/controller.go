package i18n

import (
	"gopkg.in/orivil/orivil.v0"
)

type Controller struct {
	*orivil.App
}

// SetLang for set the cookie to client, this should be a Ajax request
//
// @route {get}/setlang/::language
func (c *Controller) Setlang() {
	name := c.Params["language"]
	if shortLang, ok := Config.Languages[name]; ok {
		if name != Config.DefaultLang {
			c.SetCookie(Config.CookieKey, shortLang, Year)
		} else {
			c.SetCookie(Config.CookieKey, shortLang, 0)
		}
		c.With("success", "success")
	}
}
