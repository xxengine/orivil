package orivil

import (
	"gopkg.in/orivil/event.v0"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	LsnServer = "orivil.ServerListener"
)

type ServerListener struct {}

func (this *ServerListener) RegisterService(server *Server) {
	for _, provider := range server.Registers {
		provider.RegisterService(server.SContainer)
	}
}

func (this *ServerListener) RegisterRoute(server *Server) {
	for _, provider := range server.Registers {
		provider.RegisterRoute(server.RContainer)
	}
}

func (this *ServerListener) RegisterMiddle(server *Server) {
	for _, provider := range server.Registers {
		provider.RegisterMiddle(server.MContainer)
	}
}

func (this *ServerListener) BootProvider(server *Server) {
	for _, provider := range server.Registers {
		provider.Boot(server.SContainer)
	}
}

func (this *ServerListener) ConfigServer(s *Server) {

	actions := s.RContainer.GetActions()
	controllers := s.RContainer.GetControllers()

	// 1. add all actions name to middleware bag
	for bundle, controllers := range actions {
		for controller, actions := range controllers {
			s.MiddleBag.AddController(bundle, controller, actions)
		}
	}

	// 2. config bundle and controller middleware
	for _, provider := range s.Registers {
		if register, ok := provider.(MiddlewareConfigure); ok {
			bundle := filepath.Base(reflect.TypeOf(provider).Elem().PkgPath())
			s.MiddleBag.SetCurrent(bundle, "")
			register.SetMiddle(s.MiddleBag)
		}
	}

	// 3. config action middleware
	for name, provider := range controllers {
		bundle := name[0:strings.Index(name, ".")]
		controller := name[len(bundle)+1:]
		instance := provider()
		if register, ok := instance.(MiddlewareConfigure); ok {
			s.MiddleBag.SetCurrent(bundle, controller)
			register.SetMiddle(s.MiddleBag)
		}
	}
}

func (this *ServerListener) GetSubscribe() (name string, subscribes []event.Subscribe) {
	name = LsnServer
	p := 1000
	subscribes = []event.Subscribe{
		{
			Name:     EvtRegisterService,
			Priority: p,
		},

		{
			Name:     EvtRegisterRoute,
			Priority: p,
		},

		{
			Name:     EvtRegisterMiddle,
			Priority: p,
		},

		{
			Name:     EvtBootProvider,
			Priority: p,
		},

		{
			Name:     EvtConfigProvider,
			Priority: p,
		},
	}
	return
}
