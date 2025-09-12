package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	pb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
)

func GetRepository(ctx context.Context) *pb.SourceRepository {
	repo := uint64(*repoid)
	if repo != 0 {
		fmt.Printf("Display repo #%d\n", repo)
		rid := &pb.ByIDRequest{ID: repo}
		rl, err := pb.GetGIT2Client().RepoByID(ctx, rid)
		utils.Bail("failed to get repo", err)
		return rl
	}
	tr, err := findGitTopDir()
	utils.Bail("failed to find .git in repo", err)
	gc, err := ParseGitConfig(tr + "/.git/config")
	utils.Bail("failed to parse git config", err)
	url := gc.GetEntry(`remote "origin"`, "url")
	if Format() == FORMAT_HUMAN {
		fmt.Printf("Repository root: %s\n", tr)
		fmt.Printf("URL            : \"%s\"\n", url)
	} else if Format() == FORMAT_SHELL {
		fmt.Printf("GITSERVER_REPOSITORY_ROOT=%s\n", tr)
		fmt.Printf("GITSERVER_REPOSITORY_URL=\"%s\"\n", url)
	} else {
		panic("inv format")
	}
	if *debug {
		for _, s := range gc.Sections {
			fmt.Printf("Section: \"%s\"\n", s.Name)
			for k, v := range s.Entries {
				fmt.Printf("   Entry \"%s\": \"%s\"\n", k, v)
			}
		}
	}

	req := &pb.ByURLRequest{URL: url}
	sr, err := pb.GetGIT2Client().RepoByURL(ctx, req)
	utils.Bail(fmt.Sprintf("failed to get repo for url \"%s\"", url), err)
	return sr
}
func Latest() {
	ctx := authremote.Context()
	sr := GetRepository(ctx)
	fmt.Printf("Repository  : #%d\n", sr.ID)
	lreq := &pb.ByIDRequest{ID: sr.ID}
	build, err := pb.GetGIT2Client().GetLatestBuild(ctx, lreq)
	utils.Bail("failed to get latest build", err)
	fmt.Printf("Latest Build: %d\n", build.ID)
	fmt.Printf("LogMessage  : %s\n", build.LogMessage)
	fmt.Printf("Timestamp   : %s\n", utils.TimestampString(build.Timestamp))
	if !build.Success {
		fmt.Printf("*** build failed ***\n")
	}
}

// traverse path upwards until .git is found
func findGitTopDir() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return traverseToTopOfRepo(cur)
}

func traverseToTopOfRepo(dir string) (string, error) {
	if utils.FileExists(dir + "/.git") {
		return dir, nil
	}
	l := filepath.Dir(dir)
	if l == dir || len(l) < 2 {
		fmt.Printf("Cannot find git repository.\n")
		os.Exit(10)
	}
	return traverseToTopOfRepo(l)
}
