package main

import (
	"flag"
	"fmt"
	"golang.conradwood.net/go-easyops/linux"
	"strings"
)

type Update struct {
	ev           *Environment
	ref          string
	oldrev       string
	newrev       string
	changedFiles []*ChangedFile
}
type ChangedFile struct {
	typeLetter string
	filename   string
}

func (c *ChangedFile) ActionName() string {
	if c.typeLetter == "M" {
		return "modified"
	} else if c.typeLetter == "A" {
		return "added"
	} else {
		return "WEIRD(%s)" + c.typeLetter
	}
}

func (u *Update) Process(e *Environment) error {
	u.ev = e
	args := flag.Args()
	if len(args) != 3 {
		return fmt.Errorf("git update expected 3 args (not %d)\n", len(args))
	}
	u.ref = args[0]
	u.oldrev = args[1]
	u.newrev = args[2]
	if u.oldrev == "0000000000000000000000000000000000000000" {
		fmt.Printf("Accepting initial commit\n")
		return nil
	}
	err := u.ChangedFileNames()
	if err != nil {
		return err
	}

	if !u.ev.IsYacloudAdmin() {
		for _, f := range u.changedFiles {
			fmt.Printf("File %s: %s\n", f.ActionName(), f.filename)
			if strings.Contains(f.filename, "deployment/") {
				return fmt.Errorf("please do not modify deployment/* files")
			}
			if strings.Contains(f.filename, "autobuild.sh") {
				return fmt.Errorf("please do not modify autobuild.sh")
			}
			if strings.Contains(f.filename, "Makefile") {
				return fmt.Errorf("please do not modify any Makefile")
			}
		}
	}
	return nil
}

// git diff --name-status <sha-old> <sha-new>
func (u *Update) ChangedFileNames() error {
	var res []*ChangedFile
	l := linux.New()
	l.SetRuntime(15)
	out, err := l.SafelyExecute([]string{"git", "diff", "--name-status", u.oldrev, u.newrev}, nil)
	if err != nil {
		fmt.Printf("Git failed. Output:\n%s\n", out)
		return err
	}
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 3 {
			continue
		}
		fname := line[1:]
		fname = strings.TrimPrefix(fname, " ")
		cf := &ChangedFile{typeLetter: line[:1], filename: fname}
		res = append(res, cf)
	}
	u.changedFiles = res
	//	fmt.Printf("Changed files:\n%s\n", out)
	return nil
}
