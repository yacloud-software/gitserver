package main

import (
	"context"
	"flag"
	"fmt"
	apb "golang.conradwood.net/apis/auth"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"strconv"
	"strings"
	"time"
)

var (
	git_path        = flag.String("workdir", "/srv/temp/gitbackfillpath", "where to checkout git and stuff")
	replace_domains = map[string]string{
		"singingcat.de": "singingcat.net",
	}
)

func main() {
	flag.Parse()
	utils.Bail("failed to backfill commits: ", backfill_commits())
}
func userbyemail(email string) (*apb.User, error) {
	for k, v := range replace_domains {
		if strings.HasSuffix(email, k) {
			email = strings.TrimSuffix(email, k)
			email = email + v
			break
		}
	}

	ctx := authremote.Context()
	user, err := authremote.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func backfill_commits() error {
	ctx := context.Background()
	utils.Bail("cannot create", utils.RecreateSafely(*git_path))
	repos, err := db.DefaultDBSourceRepository().All(ctx)
	if err != nil {
		return err
	}
	for _, r := range repos {
		fmt.Printf("Repo: %d\n", r.ID)
		urls, err := db.DefaultDBSourceRepositoryURL().ByV2RepositoryID(ctx, r.ID)
		if err != nil {
			return err
		}
		if len(urls) == 0 {
			continue
		}
		url := urls[0]
		commits, err := commits_in(fmt.Sprintf("foorepo-%d", r.ID), fmt.Sprintf("https://%s/git/%s", url.Host, url.Path))
		if err != nil {
			fmt.Printf("Failed to count: %s\n", err)
			continue
		}
		for _, c := range commits.Commits() {
			fmt.Printf("%s %s %v\n", utils.TimestampString(uint32(c.Epoch)), c.Email, c.IsWrite)

			user, err := userbyemail(c.Email)
			if err != nil {
				fmt.Printf("Skipping user: %s\n", c.Email)
				continue
			}
			if user == nil {
				fmt.Printf("Skipping user: %s\n", c.Email)
				continue
			}
			gal := &gitpb.GitAccessLog{
				Write:            c.IsWrite,
				UserID:           user.ID,
				Timestamp:        uint32(c.Epoch),
				SourceRepository: r,
			}
			write(gal)
		}
	}
	return nil
}
func exists(gal *gitpb.GitAccessLog) (bool, error) {
	ctx := context.Background()
	exs, err := db.DefaultDBGitAccessLog().ByTimestamp(ctx, gal.Timestamp)
	if err != nil {
		return false, err
	}
	for _, ex := range exs {
		if ex.UserID == gal.UserID && ex.SourceRepository.ID == gal.SourceRepository.ID {
			return true, nil
		}
	}
	return false, nil

}
func write(gal *gitpb.GitAccessLog) {
	ex, err := exists(gal)
	utils.Bail("failed to check if exists", err)
	if ex {
		return
	}
	ctx := context.Background()
	_, err = db.DefaultDBGitAccessLog().Save(ctx, gal)
	utils.Bail("failed to save git access log", err)
}

type commits struct {
	commits []*Commit
}

func commits_in(reponame string, url string) (*commits, error) {
	fmt.Printf("GIT-Cloning %s into %s\n", url, reponame)
	err := exe("git", "clone", url, reponame)
	if err != nil {
		return nil, err
	}
	l := linux.New()
	l.SetMaxRuntime(time.Duration(3) * time.Minute)
	com := []string{"git", "log", "--date=unix", "--pretty=%ad %ae"}
	repodir := fmt.Sprintf("%s/%s", *git_path, reponame)
	out, err := l.SafelyExecuteWithDir(com, repodir, nil)
	if err != nil {
		fmt.Printf("Output:\n%s\n", out)
		return nil, err
	}
	//	fmt.Printf("Commits:\n", out)
	res := &commits{}
	for _, line := range strings.Split(out, "\n") {
		sx := strings.SplitN(line, " ", 2)
		if len(sx) != 2 {
			continue
		}
		ct := &Commit{}
		epoch, err := strconv.ParseUint(sx[0], 10, 64)
		if err != nil {
			fmt.Printf("Line skipped: %s\n", err)
			continue
		}
		ct.Epoch = epoch
		ct.Email = sx[1]
		ct.IsWrite = true
		res.commits = append(res.commits, ct)
	}
	err = utils.RemoveAll(repodir)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func exe(com ...string) error {
	l := linux.New()
	l.SetMaxRuntime(time.Duration(3) * time.Minute)
	out, err := l.SafelyExecuteWithDir(com, *git_path, nil)
	if err != nil {
		fmt.Printf("Output:\n%s\n", out)
		return err
	}
	return nil
}
func (c *commits) Commits() []*Commit {
	return c.commits
}

type Commit struct {
	Epoch   uint64
	Email   string
	IsWrite bool
}
