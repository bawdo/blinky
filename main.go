package main

import (
	"os"

	"github.com/bawdo/blinky/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
