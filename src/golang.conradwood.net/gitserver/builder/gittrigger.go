package builder

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/db"
	//	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/gitserver/artefacts"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"strings"
)

/*
this is the line that gets submitted via tcp from the git hook
*/

// the first line:
type GitTrigger struct {
	repodir    string
	ref        string
	oldrev     string
	newrev     string
	artefactid uint64
	foo        string
	gitinfo    *pb.GitInfo
	//	repo    *pb.Repository
}

func ParseGitTrigger(line string) (*GitTrigger, error) {
	res := &GitTrigger{}
	// parse the first line
	line = strings.TrimSuffix(line, "\r")
	line = strings.TrimSuffix(line, "\n")
	fmt.Printf("Line: \"%s\"\n", line)
	parts := strings.Split(line, " ")
	if len(parts) < 6 {
		fmt.Printf("Please update script: /usr/local/bin/git-auto-builder\n")
		return nil, fmt.Errorf("protocol error: less than 6 items (%d) received in git request (%s)\n", len(parts), line)
	}
	res.repodir = parts[0]
	res.ref = parts[1]
	res.oldrev = strings.TrimPrefix(parts[2], "x")
	res.newrev = strings.TrimPrefix(parts[3], "x")
	res.foo = strings.TrimPrefix(parts[4], "x")
	gis := strings.TrimPrefix(parts[5], "x")
	gp := &pb.GitInfo{}
	err := utils.Unmarshal(gis, gp)
	if err != nil {
		return nil, fmt.Errorf("Invalid GITINFO: %s", err)
	}
	res.gitinfo = gp

	repo, err := db.DefaultDBSourceRepository().ByID(context.Background(), res.RepositoryID())
	if err != nil {
		return nil, err
	}
	res.artefactid, err = artefacts.RepositoryIDToArtefactID(repo)
	if err != nil {
		// can't error here, because there is no artefact on first commit
		fmt.Printf("Unable to resolve repository to artefact: %s\n", utils.ErrorString(err))
	}
	return res, nil
}
func (g *GitTrigger) ExcludeBuildScripts() []string {
	return nil
}
func (g *GitTrigger) Branch() string {
	return strings.TrimPrefix(g.ref, "refs/heads/")
}

// this creates a context from g.gitinfo.User
func (g *GitTrigger) GetContext() (context.Context, error) {
	user := g.gitinfo.User
	if user == nil {
		return nil, fmt.Errorf("No user in gitinfo")
	}
	ctx, err := authremote.ContextForUserWithTimeout(user, LONG_RUNNING_SECS) // used for git clone stuff too
	if err != nil {
		return nil, err
	}
	return ctx, nil

}
func (g *GitTrigger) UserID() string {
	return g.gitinfo.UserID
}
func (g *GitTrigger) NewRev() string {
	return g.newrev
}
func (g *GitTrigger) ArtefactID() uint64 {
	return g.artefactid
}
func (g *GitTrigger) RepositoryID() uint64 {
	return g.gitinfo.RepositoryID
}



