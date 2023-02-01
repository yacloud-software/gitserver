package main

import (
	"flag"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/client"
	"golang.conradwood.net/go-easyops/linux"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"strings"
	"time"
)

/*
this runs in the GIT subprocess (not the gitserver)
*/

const (
	CALL_GITSERVER = true
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
	content    map[string]string // revision->body of file
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

	// connect to local gitserver and execute within gitserver process space
	// see readme.txt
	if CALL_GITSERVER {
		gip := "localhost:" + os.Getenv("GITSERVER_GRPC_PORT")
		fmt.Printf("grpc connection to \"%s\"\n", gip)
		con, err := client.ConnectWithIP(gip)
		if err != nil {
			return err
		}
		gc := gitpb.NewGIT2Client(con)
		srv, err := gc.RunLocalHook(e.ctx, &gitpb.HookRequest{
			RequestKey: os.Getenv("GITSERVER_KEY"),
			NewRev:     u.newrev,
			OldRev:     u.oldrev,
			HookName:   "update",
		})
		if err != nil {
			return err
		}
		for {
			hr, err := srv.Recv()
			if hr != nil {
				if hr.Output != "" {
					fmt.Print(hr.Output)
				}
				if hr.ErrorMessage != "" {
					return fmt.Errorf("%s", hr.ErrorMessage)
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}
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

	err = u.CheckFileContent()
	if err != nil {
		fmt.Printf(`


**********************************************************************************************
WARNING - files rejected in update hook
Error message is: %s

this will be rejected in future versions
**********************************************************************************************


`, err)
		// return err
	}

	return nil
}

// git diff --name-status <sha-old> <sha-new>
func (u *Update) ChangedFileNames() error {
	var res []*ChangedFile
	l := linux.New()
	l.SetMaxRuntime(time.Duration(15) * time.Second)
	fmt.Printf("Running diff to find changed filenames...\n")
	out, err := l.SafelyExecute([]string{"git", "diff", "--name-status", u.oldrev, u.newrev}, nil)
	if err != nil {
		fmt.Printf("Git diff %s %s failed. Output:\n%s\n", u.oldrev, u.newrev, out)
		return err
	}
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 3 {
			continue
		}
		fname := line[1:]
		fname = strings.TrimPrefix(fname, " ")
		fname = strings.TrimPrefix(fname, "\t")
		cf := &ChangedFile{
			typeLetter: line[:1],
			filename:   fname,
			content:    make(map[string]string),
		}
		res = append(res, cf)
	}
	u.changedFiles = res
	//	fmt.Printf("Changed files:\n%s\n", out)
	return nil
}
func (u *Update) CheckFileContent() error {
	var err error
	for _, cf := range u.changedFiles {
		if strings.HasSuffix(cf.filename, ".yaml") {
			err = u.checkYaml(cf)
		}
		if err != nil {
			return fmt.Errorf("File \"%s\" rejected by git-server checks: %s", cf.filename, err)
		}
	}
	return nil
}
func (u *Update) GetFileContent(filename string) (string, error) {
	fmt.Printf("Oldrev: %s\n", u.oldrev)
	fmt.Printf("Newrev: %s\n", u.newrev)

	//	exe("git diff " + u.oldrev + " " + u.newrev)
	gitdir := os.Getenv("GIT_BARE_REPO")
	fname := filename
	fname = strings.Trim(fname, " ")
	fmt.Printf("------ changed file (%s) -------\n", fname)
	out, err := exe("git --no-pager --git-dir " + gitdir + " show " + u.newrev + ":" + fname)
	if err != nil {
		return "", err
	}
	return out, nil
	//	return fmt.Errorf("no file content")
}
func (u *Update) checkYaml(cf *ChangedFile) error {
	fmt.Printf("Checking yaml file: %s\n", cf.filename)
	body, err := cf.GetContent(u.newrev)
	if err != nil {
		return err
	}
	foo := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(body), foo)
	if err != nil {
		return err
	}
	return nil
}

func (cf *ChangedFile) GetContent(rev string) (string, error) {
	body, found := cf.content[rev]
	if found {
		return body, nil
	}
	gitdir := os.Getenv("GIT_BARE_REPO")
	fname := cf.filename
	fname = strings.Trim(fname, " ")
	//	fmt.Printf("------ changed file (%s) -------\n", fname)
	body, err := exe("git --no-pager --git-dir " + gitdir + " show " + rev + ":" + fname)
	if err != nil {
		return "", err
	}
	cf.content[rev] = body
	return body, nil
}

// run command and return output
func exe(com string) (string, error) {
	fmt.Printf("Executing \"%s\"\n", com)
	l := linux.New()
	l.SetMaxRuntime(time.Duration(15) * time.Second)
	cmd := []string{"bash", "-c", com}
	out, err := l.SafelyExecute(cmd, nil)
	fmt.Println(out)
	if err != nil {
		return "", err
	}
	return out, nil
}
