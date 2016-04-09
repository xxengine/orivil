package orivil

type RequestHandler interface {
	Handle(app *App)
}

type TerminateHandler interface {
	Terminate(app *App)
}
