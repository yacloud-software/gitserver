package builder

import (
	"context"
	"fmt"
	"golang.conradwood.net/apis/gitbuilder"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/sql"
	"io"
	"time"
)

// use the gitbuilder service to build instead of locally forking it off
// e.g. ref== 'master', newrev == commitid
func external_builder(gt *GitTrigger, w io.Writer) error {
	ctx, err := gt.GetContext()
	if err != nil {
		return err
	}
	gi := gt.gitinfo

	psql, err := sql.Open()
	if err != nil {
		return err
	}
	// create build
	bdb := db.NewDBBuild(psql)
	nb := &gitpb.Build{
		UserID:       gi.UserID,
		RepositoryID: gi.RepositoryID,
		CommitHash:   gt.newrev,
		Branch:       gt.Branch(),
		LogMessage:   "logmessage unavailable", // to get the logmessage we have to check the repo out
		Timestamp:    uint32(time.Now().Unix()),
	}

	id, err := bdb.Save(context.Background(), nb)
	if err != nil {
		fmt.Printf("Failed to save to database: %s\n", err)
		return err
	}

	repo, err := db.NewDBSourceRepository(psql).ByID(ctx, gi.RepositoryID)
	if err != nil {
		return err
	}
	urls, err := db.NewDBSourceRepositoryURL(psql).ByV2RepositoryID(ctx, gi.RepositoryID)
	if err != nil {
		return err
	}
	if len(urls) == 0 {
		return fmt.Errorf("Repository %d (%s) has no urls\n", repo.ID, repo.ArtefactName)
	}
	url := fmt.Sprintf("https://%s/git/%s", urls[0].Host, urls[0].Path)

	gb := gitbuilder.GetGitBuilderClient()
	br := &gitbuilder.BuildRequest{
		GitURL:       url,
		CommitID:     gt.newrev,
		BuildNumber:  id,
		RepositoryID: gi.RepositoryID,
		RepoName:     repo.ArtefactName,
		ArtefactName: repo.ArtefactName,
	}
	cl, err := gb.Build(ctx, br)
	if err != nil {
		return err
	}
	var lastResponse *gitbuilder.BuildResponse
	for {
		res, err := cl.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if res.Complete {
			lastResponse = res
		}
		if len(res.Stdout) != 0 {
			_, err = w.Write(res.Stdout)
			if err != nil {
				return err
			}
		}
	}
	if !lastResponse.Success {
		return fmt.Errorf("build failed (%s)", lastResponse.ResultMessage)
	}
	return nil
}
