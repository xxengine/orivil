package orivil

import (
	"reflect"
)

// RouteFilter filter actions be registered to router
type RouteFilter struct {
	structs []interface{}
	actions []string
}

func NewRouteFilter() *RouteFilter {
	return &RouteFilter{}
}

// AddStructs filter all structs methods
func (f *RouteFilter) AddStructs(structs []interface{}) {
	f.structs = structs
}

// AddActions filter all actions
func (f *RouteFilter) AddActions(actions []string) {
	f.actions = actions
}

// implement router.ActionFilter interface for filter actions
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
