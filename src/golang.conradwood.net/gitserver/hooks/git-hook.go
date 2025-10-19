package main

/*
* this is the actual "hook" executed by git
 */
import (
	"flag"
	"fmt"
	"os"

	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/utils"
)

var (
	hook_type = flag.String("hook_type", "", "hook type, e.g. update|post-receive or so")
)

func main() {
	flag.Parse()

	var err error
	ev := Setup()
	fmt.Printf("[hook] === Git Hook started: \"%s\"\n", *hook_type)
	//	fmt.Fprintf(os.Stderr, "=== Git Hook started: \"%s\"\n", *hook_type)
	if *hook_type == "update" {
		// run before git merges into codebase (can reject)
		update := Update{}
		err = update.Process(ev)
	} else if *hook_type == "post-receive" {
		// runs after git merged the codebase
		postreceive := PostReceive{}
		err = postreceive.Process(ev)
	} else {
		req := &gitpb.HookRequest{
			RequestKey: os.Getenv("GITSERVER_KEY"),
			NewRev:     "-",
			OldRev:     "-",
			HookName:   *hook_type,
		}
		err := call_gitserver(req)
		utils.Bail(fmt.Sprintf("failed to run hook \"%s\"", *hook_type), err)
	}
	//		fmt.Fprintf(os.Stderr, "=== Git Hook Finished: \"%s\"\n", *hook_type)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Git hook %s failed: %s\n", *hook_type, err)
		fmt.Printf("Hook \"%s\" failed: %s\n", *hook_type, err)
	}
	utils.Bail(fmt.Sprintf("hook \"%s\" failed", *hook_type), err)

}
