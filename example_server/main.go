package main

import (
	"gopkg.in/orivil/orivil.v0"
	"gopkg.in/orivil/orivil/example_server/bundle/base"
)

func main() {
	server := orivil.NewServer(":8080")

	server.RegisterBundle(
		new(base.Register),
	)

	server.Run()

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
