package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"path/filepath"
)

func Latest() {
	tr, err := findGitTopDir()
	utils.Bail("failed to find .git in repo", err)
	fmt.Printf("Repository root: %s\n", tr)
	gc, err := ParseGitConfig(tr + "/.git/config")
	utils.Bail("failed to parse git config", err)
	url := gc.GetEntry(`remote "origin"`, "url")
	fmt.Printf("URL: \"%s\"\n", url)
	if *debug {
		for _, s := range gc.Sections {
			fmt.Printf("Section: \"%s\"\n", s.Name)
			for k, v := range s.Entries {
				fmt.Printf("   Entry \"%s\": \"%s\"\n", k, v)
			}
		}
	}

	ctx := authremote.Context()
	req := &pb.ByURLRequest{URL: url}
	sr, err := pb.GetGIT2Client().RepoByURL(ctx, req)
	utils.Bail(fmt.Sprintf("failed to get repo for url \"%s\"", url), err)
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
	return traverseToTopOfRepo(l)
}
