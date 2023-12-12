package coderunners

// to add new packages, import them here.
// each package registers with the registry in their "init()" function

import (
	"flag"
	"fmt"
	_ "golang.conradwood.net/gitserver/coderunners/java"
	"golang.conradwood.net/gitserver/coderunners/registry"
)

var (
	activate = flag.Bool("activate_coderunners", true, "activate the code runners to run through a repo before build")
)

// this gets called by the gitserver before starting the build
// an error aborts the commit
func RunGit(ro *registry.Opts) error {
	for _, r := range registry.All() {
		fmt.Printf("Running \"%s\"\n", r.GetName())
		err := r.Run(ro)
		if err != nil {
			fmt.Printf("%s failed: %s\n", r.GetName(), err)
		}
	}
	return nil
}



