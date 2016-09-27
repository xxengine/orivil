package main

import (
	"gopkg.in/orivil/orivil.v2"
	"gopkg.in/orivil/log.v0"
	"gopkg.in/orivil/orivil.v2/example/bundle/bundleExample"
	"gopkg.in/orivil/orivil.v2/example/bundle/debug"
)

func main() {
	server := orivil.NewServer(":8080")

	// register bundles
	server.RegisterBundle(

		new(bundleExample.Register),
		new(debug.Register),
	)

	// initialize and start server
	err := server.ListenAndServe()

	if err != nil {

		log.ErrEmergencyF("server stopped! %v", err)
	}
}
