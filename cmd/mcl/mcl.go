package main

import (
	"fmt"
	"os"

	"github.com/snugfox/mcl/cmd/mcl/app"
)

func main() {
	// rand.Seed(time.Now().UnixNano()) // rand not used in this app

	command := app.NewMCLCommand()
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
