package main

/*
this is the "gitcredentials" helper (invoked by git if git requires a username/password)
this is configured in /etc/gitconfig like so:
[credential]
        helper=/usr/local/bin/gitserver-credentials

*/

import (
	"flag"
	"os"
)

var (
	server_mode = flag.Bool("server_mode", false, "debug start in server mode for testing")
)

func main() {
	flag.Parse()
	if *server_mode {
		StartServer()
		os.Exit(0)
	}
	Client()
}
