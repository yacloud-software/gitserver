package main

import (
	"flag"
	"fmt"
	"golang.conradwood.net/gitserver/builder"
	"golang.conradwood.net/gitserver/git2"
	"golang.conradwood.net/go-easyops/utils"
)

func main() {
	var err error
	flag.Parse()
	fmt.Printf("Initialising git2server...\n")

	// start the listener for the TCP connection during build (post-receive)
	err = builder.Start()
	utils.Bail("failed to start builder...\n", err)
	err = git2.Start()
	utils.Bail("failed to start git2...\n", err)
	select {}
}
