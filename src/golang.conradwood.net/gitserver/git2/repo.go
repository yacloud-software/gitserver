package git2

import (
	"context"
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/config"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/utils"
	"strings"
)

var (
	try_other_vhosts = flag.Bool("try_other_vhosts", false, "if a repository cannot be found on the specified host, try other hostnames")
)

func (g *Repo) OnDiskPath() string {
	s := strings.Trim(g.gitrepo.FilePath, "/")
	return s
}

type Repo struct {
	gitrepo    *pb.SourceRepository
	forkedRepo *pb.SourceRepository // if forking then this is non-nil and is the source repo
}

func RepoFromURL(ctx context.Context, host string, git_url *GitURL) (*Repo, error) {
	path := git_url.RepoPath()
	surls, err := db.DefaultDBSourceRepositoryURL().ByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	var surl *pb.SourceRepositoryURL
	var surl_foohost *pb.SourceRepositoryURL
	for _, su := range surls {
		surl_foohost = su
		if strings.ToLower(su.Host) == strings.ToLower(host) {
			surl = su
			break
		}
	}
	if surl == nil && !*try_other_vhosts {
		return nil, errors.NotFound(ctx, "repository at host \"%s\", path \"%s\" not found", host, path)
	}
	if surl == nil && surl_foohost == nil {
		return nil, errors.NotFound(ctx, "repository at host \"%s\", path \"%s\" not found [not even when ignoring vhost]", host, path)
	} else if surl == nil {
		surl = surl_foohost
	}

	res := &Repo{}
	g, err := db.DefaultDBSourceRepository().ByID(ctx, surl.V2RepositoryID)
	if err != nil {
		return nil, err
	}
	err = isDeleted(ctx, g)
	if err != nil {
		return nil, err
	}
	res.gitrepo = g

	return res, nil
}
func (r *Repo) ExistsOnDisk() bool {
	s := r.AbsDirectory()
	b := utils.FileExists(s)
	if !b {
		fmt.Printf("Directory \"%s\" (repo #%d) not found\n", s, r.gitrepo.ID)
	}
	return b
}

// the absolute path to the git directory
func (r *Repo) AbsDirectory() string {
	return *config.Gitroot + "/" + r.OnDiskPath()
}


