package main

import (
	"fmt"
	_ "golang.conradwood.net/coderunners/java"
	"golang.conradwood.net/coderunners/registry"
)

func main() {
	fmt.Printf("CodeRunner...\n")
	o := &registry.Opts{}
	for _, r := range registry.All() {
		fmt.Printf("Running \"%s\"\n", r.GetName())
		err := r.Run(o)
		if err != nil {
			fmt.Printf("%s failed: %s\n", r.GetName(), err)
		}
	}
}
