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

	urls, err := targets(connections)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) {
			fmt.Fprintln(os.Stderr, "No internet connection.")
			os.Exit(1)
		}
		log.Fatal(err)
	}

	m := NewModel(urls)
	m.run()
}
