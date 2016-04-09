package orivil

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"text/template"
)

var ErrorLogger = log.New(os.Stderr, "", log.Ltime)

var tpl = template.New("error")

func init() {
	_, err := tpl.Parse(tmpl)
	if err != nil {
		panic(err)
	}
}

func CoverError(w http.ResponseWriter, r *http.Request, call func()) {
	defer func() {
		err := recover()
		if err != nil {
			if data, ok := err.(*redirect); ok {
				http.Redirect(w, r, data.url, data.code)
			} else {

				w.WriteHeader(http.StatusInternalServerError)
				if CfgApp.Debug {
					skip := 3
					var trace []string
					for {
						_, file, line, ok := runtime.Caller(skip)
						if !ok {
							break
						}
						trace = append(trace, fmt.Sprintf("%s: %d", file, line))
						skip++
					}
					tpl.Execute(w, map[string]interface{}{
						"errMsg": err,
						"trace":  trace,
					})
					// 同时将错误打印到控制台
					fmt.Println(err)
					for _, t := range trace {
						fmt.Println(t)
					}
				} else {
					w.Write([]byte("500 Server Error"))

					addr := GetIp(r)
					ErrorLogger.Printf("Ip: %s\nUrl: %s\n %v\n", addr, r.URL.String(), err)
				}
			}
		}
	}()
	call()
}

func GetIp(r *http.Request) string {

	addr := r.Header.Get("X-Real-IP")
	if addr == "" {
		addr = r.Header.Get("X-Forwarded-For")
		if addr == "" {
			addr = r.RemoteAddr
		}
	}
	return addr
}

var tmpl = `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Orivil Panic</title>
	<style type="text/css">
        .alert-danger {
            color: #a94442;
            background-color: #f2dede;
            border-color: #ebccd1;
        }
        .alert {
            padding: 15px;
            margin-bottom: 20px;
            border: 1px solid transparent;
            border-radius: 4px;
        }
        .panel {
			margin-bottom: 20px;
			background-color: #fff;
			border: 1px solid transparent;
			border-radius: 4px;
			-webkit-box-shadow: 0 1px 1px rgba(0,0,0,.05);
			box-shadow: 0 1px 1px rgba(0,0,0,.05);
		}
		.panel-info {
			border-color: #bce8f1;
		}
		.panel-info>.panel-heading {
			color: #31708f;
			background-color: #d9edf7;
			border-color: #bce8f1;
		}
		.panel-heading {
			padding: 10px 15px;
			border-bottom: 1px solid transparent;
			border-top-left-radius: 3px;
			border-top-right-radius: 3px;
		}
		.panel:last-child {
			margin-bottom: 0;
		}
		.panel>.list-group {
			margin: 0;
		}
		.list-group {
			padding-left: 0;
		}
		ul {
			margin-top: 0;
			margin-bottom: 10px;
		}
		* {
			-webkit-box-sizing: border-box;
			-moz-box-sizing: border-box;
			box-sizing: border-box;
		}
		ul {
			display: block;
			list-style-type: disc;
			-webkit-margin-before: 1em;
			-webkit-margin-after: 1em;
			-webkit-margin-start: 0px;
			-webkit-margin-end: 0px;
			-webkit-padding-start: 40px;
		}
		.panel>.list-group .list-group-item {
			border-width: 1px 0;
			border-radius: 0;
		}
		.list-group-item {
			position: relative;
			display: block;
			padding: 10px 15px;
			margin-bottom: -1px;
			background-color: #fff;
			border: 1px solid #bce8f1;
		}
		div {
			display: block;
		}
		.container {
			padding-right: 15px;
			padding-left: 15px;
			margin-right: auto;
			margin-left: auto;
		}

		a:hover, a:focus {
			color: #f75553;
			text-decoration: underline;
		}
		a:active, a:hover {
			outline: 0;
		}
		a {
			-webkit-transition: background-color ease-in-out .15s, color ease-in-out .15s, border-color ease-in-out .15s;
			transition: background-color ease-in-out .15s, color ease-in-out .15s, border-color ease-in-out .15s;
		}
		a {
			color: #a94442;
			text-decoration: none;
		}
		a {
			background: transparent;
		}
    </style>
</head>
<body>
	<div class="container">
		<div class="alert alert-danger" role="alert">
		  <strong>Error!</strong> <a title="get help?" href="orivil.com/help/{{urlquery .errMsg}}">{{html .errMsg}}</a>
		</div>
		<div>
			<div class="panel panel-info">
			  <div class="panel-heading">Trace:</div>
			  <ul class="list-group">
				{{range .trace}}
				<li class="list-group-item">{{.}}</li>
				{{end}}
			  </ul>
			</div>
		</div>
    </div>
</body>
</html>`
