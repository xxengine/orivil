// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orivil

import (
	"gopkg.in/orivil/event.v0"
)

const (
	EvtRegisterService = "event_register_service"
	EvtRegisterRoute   = "event_register_Route"
	EvtRegisterMiddle  = "event_register_Middle"
	EvtBootProvider    = "event_boot_Provider"
	EvtConfigProvider  = "event_Config_Provider"
)

type RegisterListener interface {
	RegisterService(s *Server)

	RegisterRoute(s *Server)

	RegisterMiddle(s *Server)

	BootProvider(s *Server)

	ConfigServer(s *Server)
}

var serverEvents = []*event.Event{
	{
		Name: EvtRegisterService,
		Call: func(listener interface{}, param ...interface{}) {
			server := param[0].(*Server)
			listener.(RegisterListener).RegisterService(server)
		},
	},

	{
		Name: EvtRegisterRoute,
		Call: func(listener interface{}, param ...interface{}) {
			server := param[0].(*Server)
			listener.(RegisterListener).RegisterRoute(server)
		},
	},

	{
		Name: EvtRegisterMiddle,
		Call: func(listener interface{}, param ...interface{}) {
			server := param[0].(*Server)
			listener.(RegisterListener).RegisterMiddle(server)
		},
	},

	{
		Name: EvtBootProvider,
		Call: func(listener interface{}, param ...interface{}) {
			server := param[0].(*Server)
			listener.(RegisterListener).BootProvider(server)
		},
	},

	{
		Name: EvtConfigProvider,
		Call: func(listener interface{}, param ...interface{}) {
			server := param[0].(*Server)
			listener.(RegisterListener).ConfigServer(server)
		},
	},
}
