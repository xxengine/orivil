// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orivil

import (
	"net/http"
	"runtime"
	"gopkg.in/orivil/log.v0"
	"bytes"
	"fmt"
	"net"
	"html/template"
)

var debugTpl = template.New("error")

type Trace struct {
	File string
	Line int
}

func init() {
	debugTpl.Funcs(template.FuncMap{
		"minus": func(a, b int) int {
			return a - b
		},
	})
	_, err := debugTpl.Parse(debugTemplate)
	if err != nil {
		panic(err)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, app *App, err error) {

	if err == ErrExitGorountine {
		return
	}

	w.WriteHeader(http.StatusInternalServerError)

	skip := 3
	buf := bytes.NewBuffer(nil)
	var traces []Trace
	for {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		traces = append(traces, Trace{File: file, Line: line})
		msg := fmt.Sprintf("%s: %d\n", file, line)
		buf.WriteString(msg)
		skip++
	}

	if CfgApp.DEBUG {
		//errStr := strings.Replace(err.(error).Error(), "\n", "<br>", -1)
		execErr := debugTpl.Execute(w, map[string]interface{}{
			"errMsg": err.Error(),
			"trace":  traces,
		})
		if execErr != nil {
			log.ErrEmergency(execErr)
		}
	} else {
		w.Write(internalErrorPage)
	}

	ip, e := GetIp(r)
	if e != nil {
		log.ErrWarn(e)
	}
	log.ErrEmergencyF("http panic:\n[ IP ]: \n %s \n[ URL ]: \n %s\n[ ERROR ]: \n %v\n[ TRACE ]: \n%s", ip, r.URL.String(), err, buf)
}

func GetIp(r *http.Request) (net.IP, error) {

	addr := r.Header.Get("X-Real-IP")
	if addr == "" {
		addr = r.Header.Get("X-Forwarded-For")
		if addr == "" {
			addr = r.RemoteAddr
		}
	}

	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", addr)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", addr)
	}
	return userIP, nil
}

var internalErrorPage = []byte(`<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>500 Internal Error</title>
</head>
<style>
#warp {
  position: absolute;
  width:700px;
  height:200px;
  left:50%;
  top:50%;
  margin-left:-250px;
  margin-top:-100px;
}
</style>
<body>
  <div id="warp">
  	<h1>Whoops! 500 Internal Server Error</h1>
  </div>
</body>
</html>`)

var notFoundPage = []byte(`<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>404 Not Found</title>
</head>
<style>
#warp {
  position: absolute;
  width:700px;
  height:200px;
  left:50%;
  top:50%;
  margin-left:-250px;
  margin-top:-100px;
}
</style>
<body>
  <div id="warp">
  	<h1>Whoops! 404 Page Not Found</h1>
  </div>
</body>
</html>`)

var debugTemplate = `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>500 Internal Error</title>
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
		  <strong>Error!</strong> <a title="need help?" target="_blank" href="http://orivil.com/?help={{urlquery .errMsg}}">{{.errMsg}}</a>
		</div>
		<div>
			<div class="panel panel-info">
			  <div class="panel-heading">Trace:</div>
			  <ul class="list-group">
				{{range .trace}}
				<li class="list-group-item"><a href="/{{urlquery .File}}?debug=true&line={{.Line}}#{{minus .Line 10}}" target="_blank">{{.File}}: {{.Line}}</a></li>
				{{end}}
			  </ul>
			</div>
		</div>
    </div>
</body>
</html>`

var debugFileTemplate = `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>File Explorer</title>
</head>
<body>
<table>
      <tbody>
      {{range $idx, $file := .lines}}
      <tr>
        <td id="{{add $idx 1}}" style="text-align:right">{{add $idx 1}}</td>
      	{{if (add $idx 1|eq $.line)}}
        <td id="LC1" style="background:#404040; color:#FFFFFF"><pre style="margin:0;padding-left:10px;"><code>{{$file}}</code></pre></td>
        {{else}}
        <td id="LC1"><pre style="margin:0;padding-left:10px;"><code>{{$file}}</code></pre></td>
        {{end}}
      </tr>
      {{end}}
</tbody></table>
</body>
</html>`
