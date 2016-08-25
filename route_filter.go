// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orivil

import (
	"reflect"
)

// RouteFilter filters actions be registered to router
type RouteFilter struct {
	structs []interface{}
	actions []string
}

func NewRouteFilter() *RouteFilter {
	return &RouteFilter{}
}

// AddStructs filters structs methods
func (f *RouteFilter) AddStructs(structs []interface{}) {
	f.structs = structs
}

// AddActions filters actions
func (f *RouteFilter) AddActions(actions []string) {
	f.actions = actions
}

// FilterAction used for router.NewContainer()
func (f *RouteFilter) FilterAction(action string) bool {
	for _, faction := range f.actions {
		if faction == action {
			return false
		}
	}
	for _, ints := range f.structs {
		if _, ok := reflect.TypeOf(ints).MethodByName(action); ok {
			return false
		}
	}
	return true
}
