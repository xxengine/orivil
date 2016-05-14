// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orivil

import (
	"encoding/json"
	"gopkg.in/orivil/router.v0"
	"gopkg.in/orivil/service.v0"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"unicode"
	"path/filepath"
	"gopkg.in/orivil/view.v0"
)

type FileStorage func(srcFile multipart.File, name string) error

// Check error
var UploadFileTooLarge = func(err error) bool {

	if err != nil {
		return err.Error() == "multipart: Part Read: http: request body too large"
	}
	return false
}

type App struct {
	Response         http.ResponseWriter
	Request          *http.Request
	Container        *service.Container // private container
	VContainer       *view.Container
	Params           router.Param
	Action           string             // action full name like "package.controller.index"
	query            url.Values
	form             url.Values
	viewData         map[string]interface{}
	viewBundle       string
	viewFile         string
	memorySession    Session
	permanentSession PSession
	sessionContainer *service.Container
	usedApi          bool
}

func (app *App) FormFiles(fileSize, maxMemory int64, store FileStorage) error {

	// limit file size
	app.Request.Body = http.MaxBytesReader(app.Response, app.Request.Body, fileSize)

	// limit memory size
	err := app.Request.ParseMultipartForm(maxMemory)
	if err != nil {
		return err
	}

	// collect opened files for closing them
	var openedFiles []multipart.File
	defer func() {
		for _, file := range openedFiles {
			file.Close()
		}
	}()

	// range files
	files := app.Request.MultipartForm.File
	for _, headers := range files {
		for _, header := range headers {
			file, err := header.Open()
			if err != nil {
				return err
			}

			// append to close-able collections
			openedFiles = append(openedFiles, file)

			// save the file
			err = store(file, header.Filename)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (app *App) Form() url.Values {

	if app.form == nil {
		err := app.Request.ParseForm()
		if err != nil {
			// block current http goroutine continue to execute, the server will
			// recover the error and handle it with 'orivil.Err()' method
			panic(err)
		}
		// add route params to form value
		for key, value := range app.Params {
			app.Request.PostForm.Add(key, value)
		}
		app.form = app.Request.PostForm
	}
	return app.form
}

func (app *App) Query() url.Values {

	if app.query == nil {
		app.query = app.Request.URL.Query()
		// add route params to query value
		for key, value := range app.Params {
			app.query.Add(key, value)
		}
	}
	return app.query
}

// View for store the view filename, if use default action name,
// it will set the action name's first letter to lowercase
func (app *App) View(file ...string) *App {

	if len(file) == 1 {
		app.viewFile = file[0]
		app.viewBundle = app.Action[0:strings.Index(app.Action, ".")]
	} else if len(file) == 2 {
		app.viewBundle = file[0]
		app.viewFile = file[1]
	} else {
		app.viewBundle = app.Action[0:strings.Index(app.Action, ".")]
		// use action name as file name
		app.viewFile = lowerFirstLetter(app.Action[strings.LastIndex(app.Action, ".") + 1:])
	}
	return app
}

func (app *App) With(name string, data interface{}) {

	app.viewData[name] = data
}

func (app *App) Danger(msg string) {

	app.msg(msg, "danger")
}

func (app *App) Info(msg string) {

	app.msg(msg, "info")
}

func (app *App) Success(msg string) {

	app.msg(msg, "success")
}

func (app *App) Warning(msg string) {

	app.msg(msg, "warning")
}

func (app *App) FilterI18n(msg string) (i18nMsg string) {

	if filter, ok := app.Get(SvcI18nFilter).(I18nFilter); ok {
		return filter.FilterMsg(msg)
	} else {
		return msg
	}
}

func (app *App) Redirect(url string) {

	Redirect(url)
}

func (app *App) JsonEncode(data interface{}) {

	app.Response.Header().Add("Content-Type", "application/json;charset=UTF-8")
	eco := json.NewEncoder(app.Response)
	err := eco.Encode(data)
	if err != nil {
		panic(err)
	}
}

func (app *App) AddCache(name string, service interface{}) {

	app.Container.AddCache(name, service)
}

func (app *App) GetCache(service string) interface{} {

	return app.Container.GetCache(service)
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

func (app *App) Session() Session {

	if app.memorySession == nil {
		app.memorySession = app.Container.Get(SvcMemorySession).(Session)
	}
	return app.memorySession
}

func (app *App) PSession() PSession {

	if app.permanentSession == nil {
		app.permanentSession = app.Container.Get(SvcPermanentSession).(PSession)
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

// Flash could send the file or api data to client immediately, view files can
// be sent multiple times, but api data can only be sent once
func (app *App) Flash() {
	// send view file
	if app.viewFile != "" {
		var subDir string
		if filter, ok := app.Get(SvcI18nFilter).(I18nFilter); ok {
			subDir = filter.ViewSubDir()
		}
		dir := filepath.Join(DirBundle, app.viewBundle, "view", subDir)
		err := app.VContainer.Display(app.Response, dir, app.viewFile, app.viewData)
		if err != nil {
			panic(err)
		}

		// api data can only be sent once
	} else if !app.usedApi {
		// send api data
		if len(app.viewData) > 0 {
			app.JsonEncode(app.viewData)
			app.usedApi = true
		}
	}
	// init datas
	app.viewFile = ""
	app.viewData = make(map[string]interface{}, 1)
}

// Return will flash data to client and block current http goroutine continue
// to execute
func (app *App) Return() {
	app.Flash()
	Return()
}

func (app *App) msg(msg, typ string) {
	// set message header for api
	app.Response.Header().Set("Orivil-Msg", "true")

	app.With("msg", map[string]string{
		"type":    typ,
		"content": app.FilterI18n(msg),
	})
}

func lowerFirstLetter(s string) string {
	rs := []rune(s)
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}
