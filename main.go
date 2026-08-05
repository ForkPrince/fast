package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(version)
		return
	}

	final, err := Run(NewModel())
	if err != nil {
		log.Fatal(err)
	}

	if m, ok := final.(Model); ok && m.err != nil {
		var netErr net.Error
		if errors.As(m.err, &netErr) {
			fmt.Fprintln(os.Stderr, "No internet connection.")
			os.Exit(1)
		}
		log.Fatal(m.err)
	}
}
