package builder

import (
	"fmt"
	"golang.conradwood.net/apis/gitbuilder"
	"io"
)

// use the gitbuilder service to build instead of locally forking it off
// e.g. ref== 'master', newrev == commitid
func external_builder(gt *GitTrigger, w io.Writer) error {
	ctx, err := gt.GetContext()
	if err != nil {
		return err
	}
	gi := gt.gitinfo
	gb := gitbuilder.GetGitBuilderClient()
	br := &gitbuilder.BuildRequest{
		GitURL:       gt.gitinfo.URL,
		CommitID:     gt.newrev,
		BuildNumber:  0,
		RepositoryID: gi.RepositoryID,
		RepoName:     "",
		ArtefactName: "",
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
			w.Write(res.Stdout)
		}
	}
	if !lastResponse.Success {
		return fmt.Errorf("build failed (%s)", lastResponse.ResultMessage)
	}
	return nil
}
