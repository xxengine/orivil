package debug

import (
	"fmt"
	"time"
	"bytes"
	"gopkg.in/orivil/orivil.v2"
)

const MidDebug = "debug.ViewComponent"

var history []string
var SQLs []string
var mergedHtml []byte

func AddQuery(q string) {
	if len(SQLs) >= 20 {
		SQLs = SQLs[1:]
	}
	SQLs = append(SQLs, q)
}

type ViewComponent struct {}

func (ViewComponent) Terminate(app *orivil.App) {

	app.Defer(func(){
		if len(history) >= 20 {
			history = history[1:]
		}
		h, m, s := app.Start.Clock()
		history = append(history, fmt.Sprintf(`%02d:%02d:%02d [cost time]:<span style="color:red;">%v</span> [URL]:<span style="color:green;">%s</span>`,
			h, m, s,
			time.Since(app.Start),
			app.Request.URL))

		mergedHtml, _ = orivil.GetMergedFile(app)
	})


	if orivil.ViewDebug(app, "debug", "index") {

		app.With("debugEnvironment", orivil.GetSysInfo())

		buf := bytes.NewBuffer(nil)
		app.Server.PrintInfoAt(buf)
		app.With("debugRouteAndMiddles", buf.String())
	}
}