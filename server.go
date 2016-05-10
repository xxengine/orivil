// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package orivil organize all of the server components to be one runnable server,
// and provides some useful methods.
package orivil

import (
	"fmt"
	"gopkg.in/orivil/event.v0"
	"gopkg.in/orivil/middle.v0"
	"gopkg.in/orivil/router.v0"
	"gopkg.in/orivil/service.v0"
	. "gopkg.in/orivil/session.v0"
	"gopkg.in/orivil/view.v0"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"log"
	"github.com/orivil/gracehttp"
	"time"
	"gopkg.in/orivil/validate.v0"
)

const (
	SvcApp = "orivil.App"
)

var (
// the unique key for server, Orivil will read the value from config file "app.yml"
	Key string

// Err for handle errors, every one could used it to handle error, and
// this method can be re-defined by customers
	Err = func(e error) {

		log.Println(e)
	}

// this func is no need to re-defined
	Errf = func(format string, args ...interface{}) {

		Err(fmt.Errorf(format, args...))
	}

	Emer = func(e error) {

		log.Println(e)
	}

	Emerf = func(format string, args ...interface{}) {

		Emer(fmt.Errorf(format, args...))
	}

	Warn = func(e error) {

		log.Println(e)
	}

	Warnf = func(format string, args ...interface{}) {

		Warn(fmt.Errorf(format, args...))
	}
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

// CloseAble
type CloseAble interface {
	Close()
}

type Server struct {
	SContainer      *service.Container
	MContainer      *middle.Container
	RContainer      *router.Container
	MiddleBag       *middle.Bag
	VContainer      *view.Container
	Dispatcher      *event.Dispatcher
	Registers       []Register
	fileHandler     FileHandler
	notFoundHandler NotFoundHandler
	timeOutHandler  http.Handler
	*gracehttp.Server
}

func NewServer(addr string) *Server {

	// public service container, for store "service providers"
	sContainer := service.NewPublicContainer()

	// middleware bag for config middlewares and match middlewares
	middleBag := middle.NewMiddlewareBag()

	// middleware container dependent on service container, for store
	// middlewares to service container
	mContainer := middle.NewContainer(middleBag, sContainer)

	// view compiler
	compiler := view.NewContainer(CfgApp.Debug, CfgApp.View_file_ext)

	// RouteFilter for filter controller actions to register to router
	routeFilter := NewRouteFilter()
	// filter controller extends methods to register to router
	routeFilter.AddStructs([]interface{}{
		&App{},
	})

	// filter actions to register to router
	routeFilter.AddActions([]string{
		"SetMiddle",
	})

	// route container collect all of the controller comment, and add
	// them to the router if possible
	rContainer := router.NewContainer(DirBundle, routeFilter)

	// server dispatcher, for dispatch server event when server start
	dispatcher := event.NewDispatcher()
	dispatcher.AddEvents(serverEvents)
	dispatcher.AddListener(
		new(ServerListener),
	)

	// new server
	server := &Server{
		SContainer: sContainer,
		MiddleBag:  middleBag,
		MContainer: mContainer,
		RContainer: rContainer,
		VContainer: compiler,
		Dispatcher: dispatcher,
	}

	// use the grace http server as default http server, this server could
	// be hot update
	timeOut := time.Second * time.Duration(CfgApp.Timeout)
	server.Server = gracehttp.NewServer(addr, server, timeOut, timeOut)

	// when the grace http server received 'stop signal', the current server
	// will be closed, and before that, the "bundles" must be closed first
	server.Server.AddCloseListener(server)

	// set default not found handler
	server.notFoundHandler = server

	// set default static file server handler
	server.fileHandler = server

	// register base service
	server.RegisterBundle(
		new(BaseRegister),
	)
	return server
}

// SetNotFoundHandler for handle 404 not found
func (s *Server) SetNotFoundHandler(h NotFoundHandler) {
	s.notFoundHandler = h
}

// Close for close bundle registers, this function will be auto executed
func (s *Server) Close() {

	log.Println("closing bundle register...")
	for _, reg := range s.Registers {
		if clo, ok := reg.(CloseAble); ok {
			clo.Close()
		}
	}
}

// SetFileHandler for set customer file handler
func (s *Server) SetFileHandler(h FileHandler) {
	s.fileHandler = h
}

// AddServerListener for add server event listeners
func (s *Server) AddServerListener(ls ...event.Listener) {
	s.Dispatcher.AddListener(ls...)
}

// HandleFile for judge the client whether or not to request a static file,
// this function could be replaced by customers
func (s *Server) HandleFile(r *http.Request) bool {
	return filepath.Ext(r.URL.Path) != ""
}

// ServeFile for serve static file, this function could be replaced by customers
func (s *Server) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, name)
}

// ServeHTTP for serve http request, every request goes through the function,
// include static file request
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// handle static file
	if s.fileHandler.HandleFile(r) {
		s.fileHandler.ServeFile(w, r, filepath.Join(DirStaticFile, path))
	} else {
		var app *App
		CoverError(w, r, func() {
			path = r.Method + path

			// match route
			if action, params, controller, ok := s.RContainer.Match(path); ok {

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
					viewData:  make(map[string]interface{}, 1),
				}

				// set "app" instance to private container, so the private container could
				// use "app" as service
				app.SetInstance(SvcApp, app)

				// match middleware
				middleNames := s.MContainer.Get(action)
				middles := make([]interface{}, len(middleNames))

				// get middleware instances from private container
				for index, service := range middleNames {
					middles[index] = privateContainer.Get(service)
				}

				// call middlewares
				s.callMiddles(middles, app)

				// call controller action
				value := reflect.ValueOf(controller())
				s.setControllerDependence(value, app)
				method := action[strings.LastIndex(action, ".") + 1:]
				actionFun, _ := value.Type().MethodByName(method)
				actionFun.Func.Call([]reflect.Value{value})

				// send view file or api data
				app.Flash()

				// call "Terminate" middlewares
				s.callMiddlesTerminate(middles, app)
			} else {
				s.notFoundHandler.NotFound(w, r)
			}
		})

		if app != nil {
			s.storeSession(app)
		}
	}
}

// implement NotFoundHandler interface
func (s *Server) NotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (s *Server) storeSession(a *App) {
	// if permanent session service was used, store it
	if session, ok := a.GetCache(SvcPermanentSession).(*Session); ok {

		StorePermanentSession(session)
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
		case *validate.Validate:

			var msg string
			if app.IsGet() {
				msg = mid.Valid(app.Query())
			} else {
				msg = mid.Valid(app.Form())
			}
			if msg != "" {
				app.Warning(msg)
				app.Return()
			}
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

func (s *Server) PrintMsg() {
	routeMsg := router.GetAllRouteMsg(s.RContainer)
	fmt.Println()
	fmt.Println("route message:")
	for _, msg := range routeMsg {
		fmt.Println(msg)
	}

	actions := s.RContainer.GetActions()
	middleMsg := middle.GetMiddlesMsg(s.MContainer, actions)
	fmt.Println()
	fmt.Println("middleware message:")
	for _, msg := range middleMsg {
		fmt.Println(msg)
	}
}

func (s *Server) Run() {

	// add listeners from bundle registers
	s.addServerListener(s.Registers)

	// register service
	s.Dispatcher.Trigger(EvtRegisterService, s)

	// register route
	s.Dispatcher.Trigger(EvtRegisterRoute, s)

	// register middleware
	s.Dispatcher.Trigger(EvtRegisterMiddle, s)

	// config provider
	s.Dispatcher.Trigger(EvtConfigProvider, s)

	// boot all provider
	s.Dispatcher.Trigger(EvtBootProvider, s)
}

func (s *Server) addServerListener(registers []Register) {
	for _, provider := range registers {
		if listenable, ok := provider.(ServerEventListener); ok {
			listenable.AddServerListener(s.Dispatcher)
		}
	}
}

// RegisterBundle for add bundle register to the server
func (s *Server) RegisterBundle(app ...Register) {
	s.Registers = append(s.Registers, app...)
}
