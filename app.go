package orivil

import (
	"encoding/json"
	"gopkg.in/orivil/router.v0"
	"gopkg.in/orivil/service.v0"
	"gopkg.in/orivil/session.v0"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

type App struct {
	Response         http.ResponseWriter
	Request          *http.Request
	Container        *service.Container // private container
	Params           router.Param
	Action           string // action full name like "package.controller.index"
	viewData         map[string]interface{}
	viewFile         string
	memorySession    *session.Session
	permanentSession *session.Session
	sessionContainer *service.Container
	viewSubDir       string
	currentLang      string
}

// used for I18n
func SetViewSubDir(dir string, a *App) {
	a.viewSubDir = dir
}

// used for I18n
func SetCurrentLang(lang string, a *App) {
	a.currentLang = lang
}

func (app *App) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {

	return app.Request.FormFile(key)
}

func (app *App) Form() url.Values {

	if app.Request.PostForm == nil {
		err := app.Request.ParseForm()
		if err != nil {
			log.Panic(err)
		}
	}
	return app.Request.PostForm
}

func (app *App) Query() url.Values {

	return app.Request.URL.Query()
}

// View for store the view filename, if use default action name,
// it will set the action name's first letter to lowercase
func (app *App) View(file ...string) *App {

	if app.viewFile == "" {
		var viewFile string
		if len(file) > 0 {
			viewFile = file[0]
		} else {
			// use action name as file name
			viewFile = lowerFirstLetter(app.Action[strings.LastIndex(app.Action, ".")+1:])
		}

		app.viewFile = viewFile
	}
	return app
}

func (app *App) With(name string, data interface{}) {

	app.viewData[name] = data
}

func getMsgData(msg, typ string) (data map[string]string) {
	return map[string]string{
		"type":    typ,
		"message": msg,
	}
}

func (app *App) Danger(msg string) {

	app.With("msg", getMsgData(I18n.Filter(msg, app.currentLang), "danger"))
}

func (app *App) Info(msg string) {

	app.With("msg", getMsgData(I18n.Filter(msg, app.currentLang), "info"))
}

func (app *App) Success(msg string) {

	app.With("msg", getMsgData(I18n.Filter(msg, app.currentLang), "success"))
}

func (app *App) Warning(msg string) {

	app.With("msg", getMsgData(I18n.Filter(msg, app.currentLang), "warning"))
}

func (app *App) FilterI18n(msg string) (i18nMsg string) {

	return I18n.Filter(msg, app.currentLang)
}

func (app *App) Redirect(url string) {

	Redirect(url)
}

func (app *App) JsonEncode(data interface{}) {

	app.Response.Header().Add("Content-Type", "application/json;charset=UTF-8")
	eco := json.NewEncoder(app.Response)
	err := eco.Encode(data)
	if err != nil {
		log.Panic(err)
	}
}

func (app *App) SetInstance(name string, service interface{}) {

	app.Container.SetInstance(name, service)
}

func (app *App) HasGot(service string) (interface{}, bool) {

	return app.Container.HasGot(service)
}

func (app *App) SessionContainer() *service.Container {

	if app.sessionContainer == nil {
		app.sessionContainer = app.Container.Get(SvcSessionContainer).(*service.Container)
	}
	return app.sessionContainer
}

func (app *App) Get(service string) interface{} {

	return app.Container.Get(service)
}

func (app *App) GetNew(service string) interface{} {

	return app.Container.GetNew(service)
}

func (app *App) Session() *session.Session {

	if app.memorySession == nil {
		app.memorySession = app.Container.Get(SvcMemorySession).(*session.Session)
	}
	return app.memorySession
}

func (app *App) PSession() *session.Session {

	if app.permanentSession == nil {
		app.permanentSession = app.Container.Get(SvcMemorySession).(*session.Session)
	}
	return app.permanentSession
}

func (app *App) SetCookie(key, value string, maxAge int) {

	http.SetCookie(app.Response, &http.Cookie{
		Name:   key,
		Path:   "/",
		Value:  value,
		MaxAge: maxAge,
	})
}

func (app *App) IsPost() bool {

	return app.Request.Method == "POST"
}

func (app *App) IsGet() bool {

	return app.Request.Method == "GET"
}

func (app *App) WriteString(str string) {

	app.Response.Write([]byte(str))
}

func lowerFirstLetter(s string) string {
	rs := []rune(s)
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}
