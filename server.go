// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package orivil organizes all of the server components to be one runnable server,
// and also provides some useful methods.
package orivil

import (
	"gopkg.in/orivil/middle.v0"
	"gopkg.in/orivil/router.v0"
	"gopkg.in/orivil/service.v0"
	"gopkg.in/orivil/session.v0"
	"gopkg.in/orivil/view.v0"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"gopkg.in/orivil/grace.v0"
	"io"
	"os"
	"fmt"
	"strconv"
	"net/url"
	"bufio"
	"html/template"

	// import these packages for downloading them
	_ "gopkg.in/orivil/xsrftoken.v0"
	_ "gopkg.in/orivil/validator.v0"
	_ "gopkg.in/orivil/watcher.v0"
)

const (
	VERSION = "v2.0"
)

const (
	SvcApp = "orivil.App"
	SvcServer = "orivil.Server"
)

var (
	// the unique key for server, Orivil will read the value from config file "app.yml"
	Key string
)

type FileHandler interface {
	// HandleFile to check if handle the url as static file
	HandleFile(*http.Request) bool
	// ServeFile for serve static file
	ServeFile(w http.ResponseWriter, r *http.Request, fileName string)
}

type NotFoundHandler interface {
	NotFound(w http.ResponseWriter, r *http.Request)
}

type Server struct {
	SContainer      *service.Container
	MContainer      *middle.Container
	RContainer      *router.Container
	MiddleBag       *middle.Bag
	VContainer      *view.Container
	registers       []Register
	fileHandler     FileHandler
	notFoundHandler NotFoundHandler
	*grace.GraceServer
}

func NewServer(addr string) *Server {
	readTimeOut := time.Second * time.Duration(CfgApp.READ_TIMEOUT)
	writeTimeOut := time.Second * time.Duration(CfgApp.WRITE_TIMEOUT)
	httpServer := &http.Server{
		Addr: addr,
		ReadTimeout: readTimeOut,
		WriteTimeout: writeTimeOut,
	}
	graceServer := grace.NewGraceServer(httpServer)

	// public service container, for store "service providers"
	sContainer := service.NewPublicContainer()

	// middleware bag for config middlewares and match middlewares
	middleBag := middle.NewMiddlewareBag()

	// middleware container dependent on service container, for store
	// middlewares to service container
	mContainer := middle.NewContainer(middleBag, sContainer)

	// view combiner
	combiner := view.NewContainer(CfgApp.DEBUG, CfgApp.VIEW_FILE_EXT)

	// filtering register controller actions to router
	routeFilter := NewRouteFilter()

	// filtering register controller extends methods to router
	routeFilter.AddStructs([]interface{}{
		&App{},
	})

	// filtering register actions to router
	routeFilter.AddActions([]string{
		"SetMiddle",
	})

	// route container collect all of the controller comments, and add
	// them to the router if possible
	rContainer := router.NewContainer(DirBundle, routeFilter.FilterAction)

	server := &Server{
		SContainer: sContainer,
		MiddleBag:  middleBag,
		MContainer: mContainer,
		RContainer: rContainer,
		VContainer: combiner,
		GraceServer: graceServer,
	}

	server.Handler = server


	// set default not found handler
	server.notFoundHandler = &defaultNotFoundHandler{}

	// set default static file server handler
	server.fileHandler = &defaultFileHandler{}

	// register base service
	server.RegisterBundle(
		new(BaseRegister),
	)
	return server
}

func (s *Server) ListenAndServe() error {

	s.init()

	// if the server was graceful stopped, the error will be nil.
	err := s.GraceServer.ListenAndServe()
	s.close()
	return err
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {

	s.init()

	err := s.GraceServer.ListenAndServeTLS(certFile, keyFile)
	s.close()
	return err
}

// SetNotFoundHandler sets the 404 not found handler.
func (s *Server) SetNotFoundHandler(h NotFoundHandler) {
	s.notFoundHandler = h
}

// SetFileHandler sets the static file handler.
func (s *Server) SetFileHandler(h FileHandler) {
	s.fileHandler = h
}

// ServeHTTP serves the incoming http request, every request goes through the function,
// including static file requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	path := r.URL.Path
	// handle static file
	if s.fileHandler.HandleFile(r) {
		s.serveFile(w, r, path)
	} else {

		var app *App
		defer func() {
			err := recover()
			if app != nil {
				s.storeSession(app)
				for _, f := range app.defers {
					f()
				}
			}
			if err, ok := err.(error); ok {
				handleError(w, r, app, err)
			} else if err != nil {
				panic(err)
			}
		}()

		path = r.Method + path

		// match route
		if action, params, controller, ok := s.RContainer.Match(path); !ok {

			s.notFoundHandler.NotFound(w, r)
		} else {
			// new private container
			privateContainer := service.NewPrivateContainer(s.SContainer)

			// new app
			app = &App{
				Params:    params,
				Action:    action,
				Response:  w,
				Request:   r,
				Container: privateContainer,
				VContainer: s.VContainer,
				Server: s,
				data:  make(map[string]interface{}, 1),
				Start: start,
			}

			// cache the orivil.App and orivil.Server to private container.
			app.AddCache(SvcApp, app)
			app.AddCache(SvcServer, s)

			// match middleware
			middleNames := s.MContainer.Get(action)
			middles := make([]interface{}, len(middleNames))

			// get middleware instances from private container
			for index, service := range middleNames {
				middles[index] = privateContainer.Get(service)
			}

			// call middleware
			s.callMiddles(middles, app)

			// call controller action
			value := reflect.ValueOf(controller())
			s.setControllerDependence(value, app)
			method := action[strings.LastIndex(action, ".") + 1:]
			actionFun, _ := value.Type().MethodByName(method)
			actionFun.Func.Call([]reflect.Value{value})

			// call "Terminate" middleware
			s.callMiddlesTerminate(middles, app)

			// send view file or api data
			app.flash()
		}
	}
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, urlPath string) {
	var filename string
	if CfgApp.DEBUG {
		q := r.URL.Query()
		debug := q.Get("debug")
		if debug == "true" {
			line, err := strconv.Atoi(q.Get("line"))
			if err != nil {
				panic(err)
			}
			filename, err := url.QueryUnescape(urlPath)
			if err != nil {
				panic(err)
			}
			filename = filename[1:]
			file, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			var lines []string
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				panic(err)
			}
			tpl := template.New("debug")
			tpl.Funcs(template.FuncMap{
				"add": func(a, b int) int {
					return a + b
				},
			})
			_, err = tpl.Parse(debugFileTemplate)
			if err != nil {
				panic(err)
			}
			err = tpl.Execute(w, map[string]interface{}{
				"lines": lines,
				"line": line,
			})
			if err != nil {
				panic(err)
			}
			return
		}
	}

	if strings.HasPrefix(urlPath, "/bundle-") {
		str := strings.TrimPrefix(urlPath, "/bundle-")
		firstIdx := strings.Index(str, "/")
		var bundle string
		if firstIdx > 1 && firstIdx < len(str) {
			bundle = str[:firstIdx]
			urlPath = str[firstIdx:]
		}
		filename = filepath.Join(DirBundle, bundle, "public", urlPath)
	} else {
		filename = filepath.Join(DirStaticFile, urlPath)
	}
	s.fileHandler.ServeFile(w, r, filename)
}

func (s *Server) storeSession(a *App) {
	// if permanent session service was used, store it
	if s, ok := a.GetCache(SvcPermanentSession).(*session.Session); ok {

		session.StorePermanentSession(s)
	}
}

func (s *Server) setControllerDependence(controller reflect.Value, app *App) {
	v := controller.Elem()
	len := v.NumField()
	for i := 0; i < len; i++ {
		fi := v.Field(i)
		if fi.CanSet() && fi.Type().String() == "*orivil.App" {
			fi.Set(reflect.ValueOf(app))
			break
		}
	}
}

func (s *Server) callMiddles(middles []interface{}, app *App) {
	for _, middle := range middles {
		switch mid := middle.(type) {

		case RequestHandler:

			mid.Handle(app)
		case func(*App):

			mid(app)
		case TerminateHandler:
		default:
			panic(fmt.Errorf("unkown middleware type: %v", reflect.TypeOf(middle)))
		}
	}
}

func (s *Server) callMiddlesTerminate(middles []interface{}, app *App) {
	for _, middle := range middles {
		if h, ok := middle.(TerminateHandler); ok {
			h.Terminate(app)
		}
	}
}

func (s *Server) Version() string {

	return VERSION
}

// PrintInfo prints the server information to os.Stdout.
func (s *Server) PrintInfo() {

	s.PrintInfoAt(os.Stdout)
}

// PrintInfoAt prints the server information to the param w
func (s *Server) PrintInfoAt(w io.Writer) {
	routeMsg := router.GetAllRouteMsg(s.RContainer)
	fmt.Fprintf(w, "\n[routes]:\n")
	for _, msg := range routeMsg {
		fmt.Fprintln(w, msg)
	}

	actions := s.RContainer.GetActions()
	middleMsg := middle.GetMiddlesMsg(s.MContainer, actions)
	fmt.Fprintf(w, "\n[middlewares]:\n")
	for _, msg := range middleMsg {
		fmt.Fprintln(w, msg)
	}
}

// RegisterBundle collects all bundle registers
func (s *Server) RegisterBundle(r ...Register) {
	s.registers = append(s.registers, r...)
}

// Initialize all bundles
func (s *Server) init() {

	// register services
	for _, r := range s.registers {
		r.RegService(s.SContainer)
	}

	// register routes
	for _, r := range s.registers {
		r.RegRoute(s.RContainer)
	}


	// register middleware
	for _, r := range s.registers {
		r.RegMiddle(s.MContainer)
	}

	allActions := s.RContainer.GetActions()
	for bundle, controllers := range allActions {
		for controller, actions := range controllers {
			s.MiddleBag.AddController(bundle, controller, actions)
		}
	}

	// config middleware
	for _, r := range s.registers {

		bundle := filepath.Base(reflect.TypeOf(r).Elem().PkgPath())
		s.MiddleBag.SetCurrent(bundle, "")
		r.CfgMiddle(s.MiddleBag)
	}

	cProviders := s.RContainer.GetControllers()
	for bundle, controllers := range allActions {
		for controller, _ := range controllers {
			c := cProviders[bundle + "." + controller]()
			s.MiddleBag.SetCurrent(bundle, controller)
			if r, ok := c.(MiddlewareConfigure); ok {
				r.CfgMiddle(s.MiddleBag)
			}
		}
	}

	// boot services
	for _, r := range s.registers {
		r.Boot(s)
	}
}

func (s *Server) close() {
	for _, r := range s.registers {
		r.Close()
	}
}

// defaultFileHandler implements "FileHandler" interface for handling static files.
type defaultFileHandler struct{}

// HandleFile checks whether or not to handle the request as static file request.
func (s *defaultFileHandler) HandleFile(r *http.Request) bool {
	return filepath.Ext(r.URL.Path) != ""
}

// ServeFile serves static file
func (s *defaultFileHandler) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, name)
}

// implements NotFoundHandler interface
type defaultNotFoundHandler struct{}

func (h *defaultNotFoundHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(404)
	w.Write(notFoundPage)
}
