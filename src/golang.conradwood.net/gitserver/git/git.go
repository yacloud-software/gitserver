package git

import (
	"context"
	"flag"
	"fmt"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	git_dir = flag.String("git_dir", "/tmp/gitserver/builds", "directory where to run the builds in")
	gitlock sync.Mutex
	gits    []*Git
	basedir string
)

type Git struct {
	ID         int
	inuse      bool
	checkedout string // commitid of current checkout
	dir        string // where is it ;), e.g. /tmp/gitserver/foo/1.git
	url        string
	broken     bool // true if a git commit failed
	LogMessage string
}

func GitRepo(url string) *Git {
	basedir = "./temp"
	if *git_dir == "." {
		bdir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Failed to get current directory: %s\n", err)
		} else {
			basedir = bdir + "/temp"
		}
	} else {
		basedir = *git_dir + "/temp"
	}
	bdir, err := filepath.Abs(basedir)
	if err != nil {
		fmt.Printf("Failed to convert basedir: %s\n", err)
	} else {
		basedir = bdir
	}

	gitlock.Lock()
	defer gitlock.Unlock()
	if len(gits) == 0 {
		os.RemoveAll(basedir)
	}
	var gs *Git
	for _, g := range gits {
		if g.inuse || g.broken {
			continue
		}
		if g.url != url {
			continue
		}
		if !utils.FileExists(g.FullDirectoryPath()) {
			continue
		}
		gs = g
		break
	}
	if gs == nil {
		gs = &Git{url: url, ID: len(gits) + 1}
		gs.dir = fmt.Sprintf("%d", gs.ID)
		gits = append(gits, gs)
	}
	gs.inuse = true
	return gs

}
func (g *Git) Close() {
	g.inuse = false
}
func needDir(dir string) {
	os.MkdirAll(dir, 0777)
}

// e.g. /tmp/gitserver/builds/1
func (g *Git) FullDirectoryPath() string {
	return basedir + "/" + g.dir
}

func (g *Git) Checkout(ctx context.Context, ref string, commit string) error {
	needDir(basedir)
	fmt.Printf("Checking out %s to directory %s\n", g.url, basedir)
	g.broken = true // if we exit prematurely, it'll be broken
	l := linux.NewWithContext(ctx)
	l.SetRuntime(60)
	l.SetAllowConcurrency(true)
	if g.checkedout == "" {
		// git clone...
		s, err := l.SafelyExecuteWithDir([]string{"git", "clone", "--recurse-submodules", g.url, g.dir}, basedir, nil)
		if err != nil {
			fmt.Printf("Git said: %s\n", s)
			return fmt.Errorf("failed to clone git (%s)", err)
		}
		fmt.Printf("Git clone successful\n")
	} else {
		g.checkedout = ""
		l.SafelyExecuteWithDir([]string{"git", "reset", "--hard"}, g.FullDirectoryPath(), nil)
		s, err := l.SafelyExecuteWithDir([]string{"git", "checkout", "master"}, g.FullDirectoryPath(), nil)
		if err != nil {
			fmt.Printf("Git said: %s\n", s)
			return fmt.Errorf("failed to checkout git (%s)", err)
		}

		s, err = l.SafelyExecuteWithDir([]string{"git", "pull"}, g.FullDirectoryPath(), nil)
		if err != nil {
			fmt.Printf("Git said: %s\n", s)
			return fmt.Errorf("failed to pull git (%s)", err)
		}

		s, err = l.SafelyExecuteWithDir([]string{"git", "submodule", "update", "--init", "--recursive"}, g.FullDirectoryPath(), nil)
		if err != nil {
			fmt.Printf("Git said: %s\n", s)
			return fmt.Errorf("failed to git update submodules (%s)", err)
		}

		s, err = l.SafelyExecuteWithDir([]string{"git", "submodule", "update", "--remote"}, g.FullDirectoryPath(), nil)
		if err != nil {
			fmt.Printf("Git said: %s\n", s)
			return fmt.Errorf("failed to pull git submodule (%s)", err)
		}

	}

	s, err := l.SafelyExecuteWithDir([]string{"git", "checkout", commit}, g.FullDirectoryPath(), nil)
	if err != nil {
		fmt.Printf("Git said: %s\n", s)
		return fmt.Errorf("failed to checkout git (%s)", err)
	}
	gitlog, err := l.SafelyExecuteWithDir([]string{"git", "log", "-1"}, g.FullDirectoryPath(), nil)
	if err != nil {
		fmt.Printf("Git said: %s\n", s)
		return fmt.Errorf("failed to get git log (%s)", err)
	}

	sep := strings.Split(gitlog, "\n")
	for i, l := range sep {
		if l == "" {
			g.LogMessage = strings.Join(sep[i+1:], "\n")
			g.LogMessage = strings.TrimSuffix(g.LogMessage, "\n")
			g.LogMessage = strings.TrimSpace(g.LogMessage)
			fmt.Printf("Logmessage: \"%s\"\n", g.LogMessage)
			break
		}
		//fmt.Printf("%d. \"%s\"\n", i, l)
	}

	g.broken = false
	g.checkedout = commit

	return nil
}
