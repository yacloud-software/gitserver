package git2

import (
	"flag"
	"fmt"
	"golang.conradwood.net/go-easyops/utils"
	"golang.org/x/sys/unix"
	"io/ioutil"
)

var (
	override_githook_location = flag.String("override_githook_location", "", "if set, this is the location for the binary git-hook")
)

// install githoooks. called on each request (so we always use up-to-date hooks)
func (h *HTTPRequest) fixHooks() error {
	err := h.writeHook(h.repo.gitrepo.RunPostReceive, "post-receive")
	if err != nil {
		return err
	}
	err = h.writeHook(h.repo.gitrepo.RunPreReceive, "update")
	return err
}

/*
 we write a script here to
 1) set the hook_type to the name of hook being executed
 2) set path to current gitserver path
 TODO: use symlink (and os.Args[0]) instead?
*/
func (h *HTTPRequest) writeHook(enable bool, name string) error {
	hook_binaries := []string{
		fmt.Sprintf("%s/dist/linux/amd64/git-hook", h.pwd()),
		"/home/cnw/go/bin/git-hook",
	}

	hook_binary := ""
	for _, hb := range hook_binaries {
		if utils.FileExists(hb) {
			hook_binary = hb
			break
		}
	}
	if hook_binary == "" {
		for _, hb := range hook_binaries {
			fmt.Printf("Possible Location: \"%s\"\n", hb)
		}
		panic("Hook binary (git-hook) not found")
	}
	var content string

	content = fmt.Sprintf("#!/bin/sh\nexec %s -hook_type=%s $@\n", hook_binary, name)
	if !enable {
		content = "#!/bin/sh"
	}
	dir := h.repo.AbsDirectory()
	if dir == "" {
		return fmt.Errorf("cannot write hook - Missing gitdirectory")
	}
	if !utils.FileExists(dir) {
		return fmt.Errorf("cannot write hook - gitdirectory \"%s\" does not exist", dir)
	}
	filename := dir + "/hooks/" + name
	unix.Umask(000)
	err := ioutil.WriteFile(filename, []byte(content), 0777)
	if err != nil {
		return fmt.Errorf("unable to write hook \"%s\" (%s)", name, err)
	}
	return nil
}
