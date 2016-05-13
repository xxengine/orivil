package main

import (
	"gopkg.in/orivil/orivil.v1"
	"gopkg.in/orivil/orivil.v1/example_server/bundle/base"
	"log"
)

func main() {
	server := orivil.NewServer(":8080")

	server.RegisterBundle(
		new(base.Register),
	)

	server.Run()

	server.PrintMsg()

	err := server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
