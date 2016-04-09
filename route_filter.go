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

// AddStructs 排除所有 structs 中所有的方法
func (f *RouteFilter) AddStructs(structs []interface{}) {
	f.structs = structs
}

// AddActions 排除方法名
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
