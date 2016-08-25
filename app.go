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
	"errors"
	"time"
)

// FileStorage defines the callback which how to store uploaded files.
type FileStorage func(srcFile multipart.File, name string) error

// Check error
var ErrUploadFileTooLarge = errors.New("upload file too large")

var ErrExitGorountine = errors.New("exit current gorountine")

type App struct {
	Response         http.ResponseWriter
	Request          *http.Request
	Container        *service.Container // private container
	VContainer       *view.Container
	Params           router.Param
	Action           string             // looks like "bundle.Controller.Action"
	Start            time.Time
	Server           *Server
	query            url.Values
	form             url.Values
	data             map[string]interface{}
	viewPages        []view.Page
	defers           []func()
	memorySession    Session
	permanentSession PSession
	sessionContainer *service.Container
}

func GetViewPages(a *App) (ps []view.Page) {
	for _, p := range a.viewPages {
		if !p.Debug {
			ps = append(ps, p)
		}
	}
	return
}

func GetAllViewPages(a *App) (ps []view.Page) {

	return a.viewPages
}

func GetMergedFile(a *App) (file []byte, err error) {

	pages := GetViewPages(a)
	return a.Get(SvcServer).(*Server).VContainer.Combine(pages...)
}

// FormFiles reads and stores upload files.
//
// Usage:
//
// Step 1: define the 1st & 2nd parameters:
//
// var maxFileSize int64 = 600 << 10 // 600KB
// var maxMemorySize int64 = 60 << 20 // 60MB
//
// Step 2: define the 3rd parameter(a custom function for storing the upload files):
//
// var dir = "./uploads"
// var store orivil.FileStorage = func (srcFile multipart.File, name string) error {
//
//	 fileName := filepath.Join(dir, name)
//
//	 dstFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
//	 if err != nil {
//		 return err
//	 }
//	 defer dstFile.Close()
//
//	 _, err = io.Copy(dstFile, srcFile)
//	 if err != nil {
//		 return err
//	 }
//	 return nil
// }
//
// Step 3: check errors:
//
//	err := app.FormFiles(maxFileSize, maxMemorySize, store)
//
//	if orivil.ErrUploadFileTooLarge == err {
//
//		fmt.Println("upload file too large")
//	} else if err != nil {
//
//		fmt.Printf("upload file got error: %v", err)
//	}
func (app *App) FormFiles(maxFile, maxMemory int64, store FileStorage) error {

	// limit file size
	app.Request.Body = http.MaxBytesReader(app.Response, app.Request.Body, maxFile)

	// limit memory size
	err := app.Request.ParseMultipartForm(maxMemory)
	if err != nil {
		if err.Error() == "multipart: Part Read: http: request body too large" {
			return ErrUploadFileTooLarge
		} else {
			return err
		}
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

// Form reads the cache or parses values from three sources:
// raw query in the URL;
// parameters from router;
// the request body if request method is POST, PUT or PATCH;
func (app *App) Form() url.Values {

	if app.form == nil {
		err := app.Request.ParseForm()
		if err != nil {
			panic(err)
		}
		app.form = app.Request.PostForm

		// add route params to form value
		for key, value := range app.Params {

			app.form.Add(key, value)
		}
	}
	return app.form
}

// Query reads the cache or parses values from two sources:
// raw query in the URL;
// parameters from router;
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

// View accepts 0-1 parameters, it reads the view file under the bundle
// directory which the router matched.
//
// Most time we use this function in controllers, if you want to send
// view file in middleware, you should use ViewBundle to set which bundle
// directory you want to read.
//
// If given:
// 0 param: will use the lowercase action name as the view file name.
// 1 param: will use the param as the view file name.
func (app *App) View(arg ...string) *App {
	var file, bundle string
	switch len(arg) {
	case 1:
		file = arg[0]
	default:
		// use action name as file name
		file = lowerFirstLetter(app.Action[strings.LastIndex(app.Action, ".") + 1:])
	}
	bundle = app.Action[0:strings.Index(app.Action, ".")] // app.Action looks like "bundle.Controller.Action"
	return app.view(bundle, file, false)
}

// ViewBundle reads the view file under the given bundle directory.
func (app *App) ViewBundle(bundle, file string) *App {

	return app.view(bundle, file, false)
}

// ViewDebug will send the view file, but when pares files got error, the debug file
// will be ignored from file trace message.
// This function is built for debug component, it should be used in last middleware,
// because the debug page will be sent only if normal view page has been set.
func ViewDebug(app *App, bundle, file string) (ok bool) {

	ok = false
	for _, p := range app.viewPages {
		if !p.Debug {
			ok = true
			break
		}
	}
	if ok {
		app.view(bundle, file, true)
	}
	return
}

func (app *App) view(bundle, file string, debug bool) *App {
	var page view.Page
	var subDir string
	// get i18n view directory
	if filter, ok := app.Get(SvcI18nFilter).(I18nFilter); ok {
		subDir = filter.ViewSubDir()
	}
	dir := filepath.Join(DirBundle, bundle, "view", subDir)
	if debug {
		page = view.NewDebugPage(dir, file)
	} else {
		page = view.NewPage(dir, file)
	}
	app.viewPages = append(app.viewPages, page)
	return app
}

var ErrDataExist = errors.New("view data alreay exist")

func (app *App) With(name string, data interface{}) (err error) {

	if _, ok := app.data[name]; ok {
		return ErrDataExist
	}
	app.data[name] = data
	return nil
}

func (app *App) Danger(msg string) {

	app.Msg(msg, "danger")
}

func (app *App) Info(msg string) {

	app.Msg(msg, "info")
}

func (app *App) Success(msg string) {

	app.Msg(msg, "success")
}

func (app *App) Warning(msg string) {

	app.Msg(msg, "warning")
}

func (app *App) FilterI18n(msg string) (i18nMsg string) {

	if filter, ok := app.Get(SvcI18nFilter).(I18nFilter); ok {
		return filter.FilterMsg(msg)
	} else {
		return msg
	}
}

// Redirect sends redirect header to client, then to terminate the HTTP goroutine.
func (app *App) Redirect(url string, code ...int) {

	c := 302
	if len(code) != 0 {
		c = code[0]
	}
	http.Redirect(app.Response, app.Request, url, c)
	panic(ErrExitGorountine)
	// use panic because runtime.Goexit() couldn't work well under windows
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

func (app *App) Defer(f func()) {

	app.defers = append(app.defers, f)
}

func (app *App) flash() {
	// send view file
	if app.viewPages != nil {
		err := app.VContainer.Display(app.Response, app.data, app.viewPages...)
		if err != nil {
			panic(err)
		}
	} else {
		// send api data
		if len(app.data) > 0 {
			app.JsonEncode(app.data)
		}
	}
}

func (app *App) Msg(msg, typ string) {
	// Set message header to client for handling the response data as message data.
	app.Response.Header().Set("Orivil-Msg", typ)

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