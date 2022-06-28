package checks

import (
	"context"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/config"
	"golang.conradwood.net/go-easyops/linux"
	"gopkg.in/yaml.v2"
	//	"os"
	"flag"
	"strings"
)

var (
	run_checks = flag.Bool("run_checks", true, "if false, do not run server-space checks during hooks")
)

type checker struct {
	out          CheckOutput
	ctx          context.Context
	repo         *repository
	oldrev       string
	newrev       string
	changedFiles []*ChangedFile
}
type repository struct {
	SourceRepo *gitpb.SourceRepository
}
type ChangedFile struct {
	checker    *checker
	typeLetter string
	filename   string
	content    map[string]string // revision->body of file
}
type CheckOutput interface {
	Send(*gitpb.HookResponse) error
}

func NewChecker(ctx context.Context, repo *gitpb.SourceRepository, oldrev, newrev string) *checker {
	res := &checker{
		ctx:    ctx,
		repo:   &repository{SourceRepo: repo},
		oldrev: oldrev,
		newrev: newrev,
	}
	return res
}

func (c *checker) OnUpdate(out CheckOutput) error {
	c.out = out
	c.Printf("started onupdate()\n")
	if !*run_checks {
		c.Printf("Checks disabled\n")
		return nil
	}
	//	c.exe("env|sort")
	err := c.readChangedFiles()
	if err != nil {
		return err
	}
	for _, cf := range c.changedFiles {
		c.Printf("changed file: \"%s\"\n", cf.filename)
	}
	err = c.CheckFileContent()
	if err != nil {
		return err
	}
	return nil
}
func (c *checker) CheckFileContent() error {
	var err error
	for _, cf := range c.changedFiles {
		if strings.HasSuffix(cf.filename, ".yaml") {
			err = c.checkYaml(cf)
		}
		if err != nil {
			return fmt.Errorf("File \"%s\" rejected by git-server checks: %s", cf.filename, err)
		}
	}
	return nil
}
func (c *checker) checkYaml(cf *ChangedFile) error {
	c.Printf("Checking yaml file: %s\n", cf.filename)
	body, err := cf.GetContent(c.newrev)
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

func (c *checker) Printf(format string, args ...interface{}) {
	if c.out == nil {
		fmt.Printf(format, args...)
		return
	}
	s := fmt.Sprintf(format, args...)
	hr := &gitpb.HookResponse{Output: "[srvhook] " + s}
	c.out.Send(hr)
}

// fills checker with the changed files list
func (c *checker) readChangedFiles() error {
	var res []*ChangedFile
	l := linux.New()
	l.SetRuntime(15)
	out, err := l.SafelyExecuteWithDir([]string{"git", "diff", "--name-status", c.oldrev, c.newrev}, c.GitBareDir(), nil)
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
		fname = strings.TrimPrefix(fname, "\t")
		cf := &ChangedFile{
			checker:    c,
			typeLetter: line[:1],
			filename:   fname,
			content:    make(map[string]string),
		}
		res = append(res, cf)
	}
	c.changedFiles = res
	return nil
}

// run command and return output
func (c *checker) exe(com string) (string, error) {
	fmt.Printf("Executing \"%s\"\n", com)
	l := linux.New()
	l.SetRuntime(15)
	cmd := []string{"bash", "-c", com}
	out, err := l.SafelyExecuteWithDir(cmd, c.GitBareDir(), nil)
	if err != nil {
		c.Printf("%s", out)
		return "", err
	}
	return out, nil
}

func (cf *ChangedFile) GetContent(rev string) (string, error) {
	body, found := cf.content[rev]
	if found {
		return body, nil
	}
	gitdir := cf.checker.GitBareDir()
	fname := cf.filename
	fname = strings.Trim(fname, " ")
	//	fmt.Printf("------ changed file (%s) -------\n", fname)
	body, err := cf.checker.exe("git --no-pager --git-dir " + gitdir + " show " + rev + ":" + fname)
	if err != nil {
		return "", err
	}
	cf.content[rev] = body
	return body, nil
}

// path to git bare dir, e.g. /srv/git/byid/365
func (c *checker) GitBareDir() string {
	return *config.Gitroot + "/" + c.repo.SourceRepo.FilePath

}
