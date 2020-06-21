package main

import (
	"log"

	"github.com/snugfox/mcl/cmd/mcl/app"
)

func main() {
	// rand.Seed(time.Now().UnixNano()) // rand not used in this app
	log.SetPrefix("MCL ")

	command := app.NewMCLCommand()
	if err := command.Execute(); err != nil {
		log.Fatalln("encountered an error:", err)
	}
}
