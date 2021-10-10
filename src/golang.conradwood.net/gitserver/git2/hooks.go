package git2

import (
	"fmt"
	"golang.conradwood.net/go-easyops/utils"
	"golang.org/x/sys/unix"
	"io/ioutil"
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
	var content string

	content = fmt.Sprintf("#!/bin/sh\nexec %s/dist/linux/amd64/git-hook -hook_type=%s $@\n", h.pwd(), name)
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
