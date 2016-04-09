package orivil

import (
	"fmt"
	"github.com/orivil/event"
	"github.com/orivil/middle"
	"github.com/orivil/router"
	"github.com/orivil/service"
	. "github.com/orivil/session"
	"github.com/orivil/view"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	SvcApp = "orivil.App"
)

var (
	// the unique key for server
	Key string
)

type FileHandler interface {
	// HandleFile to check if handle the url as static file
	HandleFile(url string) bool
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
	Dispatcher      *event.Dispatcher
	Registers       []Register
	fileHandler     FileHandler
	notFoundHandler NotFoundHandler
	timeOutHandler  http.Handler
	*http.Server
}

func NewServer(addr string) *Server {

	// public service container 公共服务容器 , 主要用于对 service
	// provider 的存储, 如果从此容器中获取 service, 则该 service
	// 存在数据竞争, 每次 http 请求还会产生一个私有容器, 用于获取
	// service, 从私有容器中获取的 service 是数据安全的, 不必担心私
	// 有容器和公共容器该用在什么场合, 当你存 service 的时候自动存入
	// public container 公共容器, 取 service 的时候自动去 private
	// container 中取
	sContainer := service.NewPublicContainer()

	// middleware bag 用于中间件的配置及匹配
	middleBag := middle.NewMiddlewareBag()

	// middleware container 中间件容器依赖于服务容器, 存中间件服务时用
	// 公共容器, 取中间件服务时用私有容器
	mContainer := middle.NewContainer(middleBag, sContainer)

	// view compiler
	compiler := view.NewContainer(CfgApp.Debug, CfgApp.View_file_ext)

	// route filter 排除 controller 的 action 被注册进路由
	routeFilter := NewRouteFilter()
	// 排除 controller 继承的方法, every controller should extend App struct
	routeFilter.AddStructs([]interface{}{
		&App{},
	})
	// 排除方法名
	routeFilter.AddActions([]string{
		"SetMiddle",
	})

	// route container collect all of the controller comment,
	// add the then to the router if possible
	rContainer := router.NewContainer(DirBundle, routeFilter)

	// server dispatcher, only dispatch server event when server start
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

	// TODO:
	// time out handler
	//outTime := time.Duration(CfgApp.Timeout) * time.Second
	//timeOutHandler := http.TimeoutHandler(server, outTime, "")
	//server.Server = &http.Server{Addr: addr, Handler: timeOutHandler}
	server.Server = &http.Server{Addr: addr, Handler: server}

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

func (s *Server) SetNotFoundHandler(h NotFoundHandler) {
	s.notFoundHandler = h
}

func (s *Server) SetFileHandler(h FileHandler) {
	s.fileHandler = h
}

func (s *Server) AddServerListener(ls ...event.Listener) {
	s.Dispatcher.AddListener(ls...)
}

func (s *Server) HandleFile(url string) bool {
	return filepath.Ext(url) != ""
}

func (s *Server) ServeFile(w http.ResponseWriter, r *http.Request, name string) {
	http.ServeFile(w, r, name)
}

// ServeHTTP the http serve handler, every request goes through the function
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// handle static file
	if s.fileHandler.HandleFile(path) {
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
					viewData:  make(map[string]interface{}, 1),
				}
				app.SetInstance(SvcApp, app)

				// match middleware, new middleware and cache them in the
				// private service container
				middleNames := s.MContainer.Get(action)
				middles := make([]interface{}, len(middleNames))

				// get middleware instances from private container
				index := 0
				for _, service := range middleNames {
					middles[index] = privateContainer.Get(service)
					index++
				}

				// call middlewares
				s.callMiddles(middles, app)

				// call controller action
				value := reflect.ValueOf(controller())
				s.setControllerDependence(value, app)
				method := action[strings.LastIndex(action, ".")+1:]
				actionFun, _ := value.Type().MethodByName(method)
				actionFun.Func.Call([]reflect.Value{value})

				// send view file or api data
				s.send(app)

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

func (s *Server) send(a *App) {
	// send view file
	if len(a.viewFile) > 0 {
		bundle := a.Action[0:strings.Index(a.Action, ".")]
		// a.viewFile may contains sub dir like "/admin/login.tpl"
		dir := filepath.Join(DirBundle, bundle, "view", a.viewSubDir)
		err := s.VContainer.Display(a.Response, dir, a.viewFile, a.viewData)
		if err != nil {
			panic(err)
		}
	} else {
		// send api data
		if len(a.viewData) > 0 {
			a.JsonEncode(a.viewData)
		}
	}
}

func (s *Server) storeSession(a *App) {
	// if permanent session service was used, store it
	if inst, ok := a.HasGot(SvcPermanentSession); ok {
		session := inst.(*Session)
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
		if requestHandler, ok := middle.(RequestHandler); ok {

			requestHandler.Handle(app)
		} else if call, ok := middle.(func(*App)); ok {

			call(app)
		}
	}
}

func (s *Server) callMiddlesTerminate(middles []interface{}, app *App) {
	for _, middle := range middles {
		if requestHandler, ok := middle.(TerminateHandler); ok {
			requestHandler.Terminate(app)
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
	// add listeners from provider registered
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

func (s *Server) RegisterBundle(app ...Register) {
	s.Registers = append(s.Registers, app...)
}
