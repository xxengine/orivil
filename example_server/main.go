package main

import (
	"github.com/orivil/orivil"
	"github.com/orivil/orivil/example_server/bundle/base"
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
