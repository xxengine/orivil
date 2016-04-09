package i18n

import (
	"gopkg.in/orivil/orivil.v0"
	"net/http"
	"path/filepath"
)

const (
	// middleware name
	MidDataSender     = "i18n.DataSender"
	MidViewFileReader = "i18n.ViewFileReader"

	// cookie max age
	Year = 60 * 60 * 24 * 365
)

// DataSender 发送当前语言及所有语言数据至模板
func DataSender(a *orivil.App) {
	var shortName string
	if c, err := a.Request.Cookie(Config.CookieKey); err == http.ErrNoCookie {
		shortName = Config.Languages[Config.DefaultLang]
	} else {
		shortName = c.Value
	}
	currentLang := GetFullName(shortName)
	a.With("currentLang", currentLang)
	a.With("i18nlangs", Config.Languages)
}

// ViewDirReader 自动读取当前语言并设置 view 文件的目录
func ViewDirReader(a *orivil.App) {
	var shortName string

	// read language from cookie
	if c, err := a.Request.Cookie(Config.CookieKey); err == http.ErrNoCookie {
		shortName = Config.Languages[Config.DefaultLang]
	} else {
		shortName = c.Value
	}

	// set current language
	orivil.SetCurrentLang(shortName, a)
	if shortName != Config.Languages[Config.DefaultLang] {
		// for read i18n files
		orivil.SetViewSubDir(filepath.Join(i18nDir, shortName), a)
	}
}
