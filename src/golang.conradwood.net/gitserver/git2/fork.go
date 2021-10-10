package git2

import (
	"context"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/http"
	"golang.conradwood.net/go-easyops/utils"
	"strings"
	"sync"
	"time"
)

var (
	forklock sync.Mutex
)

func (g *GIT2) Fork(ctx context.Context, req *gitpb.ForkRequest) (*gitpb.SourceRepository, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "please authenticate")
	}
	if req.URL == nil {
		return nil, errors.InvalidArgs(ctx, "missing url", "missing url")
	}
	if req.URL.Host == "" {
		return nil, errors.InvalidArgs(ctx, "missing host", "missing host")
	}
	if strings.HasPrefix(req.URL.Path, "/git/") {
		return nil, errors.InvalidArgs(ctx, "url path must not not start with '/git/' ", "url path must not not start with '/git/' ")
	}
	if req.URL.Path == "" {
		return nil, errors.InvalidArgs(ctx, "missing path", "missing path")
	}
	if req.ArtefactName == "" {
		return nil, errors.InvalidArgs(ctx, "missing artefactname", "missing artefactname")
	}
	err := checkValidHost(ctx, req.URL.Host)
	if err != nil {
		return nil, err
	}
	bir := &gitpb.ByIDRequest{ID: req.RepositoryID}
	sr, err := g.RepoByID(ctx, bir)
	if err != nil {
		return nil, err
	}
	if sr.Deleted {
		return nil, errors.NotFound(ctx, "source repo does not exist", "source repo %d has been deleted", sr.ID)
	}
	if !sr.CreatedComplete {
		return nil, errors.FailedPrecondition(ctx, "source repository not ready", "source repository %d not created completely yet", sr.ID)
	}
	// save it again
	sr.ArtefactName = req.ArtefactName
	sr.FilePath = utils.RandomString(20)
	sr.Forking = true
	sr.ForkedFrom = sr.ID
	sr.ID = 0
	sr.UserCommits = 0
	sr.URLs = sr.URLs[:0] // remove urls
	if req.Description == "" {
		sr.Description = "fork of " + sr.Description
	} else {
		sr.Description = req.Description
	}
	_, err = g.repo_store.Save(ctx, sr)
	if err != nil {
		return nil, err
	}
	sr.FilePath = fmt.Sprintf("byid/%d", sr.ID)
	err = g.repo_store.Update(ctx, sr)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(req.URL.Path, ".git") {
		req.URL.Path = req.URL.Path + ".git"
	}
	req.URL.V2RepositoryID = sr.ID
	_, err = g.url_store.Save(ctx, req.URL)
	if err != nil {
		return nil, err
	}
	// add the create repo log
	ctxs, err := auth.SerialiseContextToString(ctx)
	if err != nil {
		return nil, err
	}

	crl := &gitpb.CreateRepoLog{
		RepositoryID:     sr.ID,
		UserID:           u.ID,
		Context:          ctxs,
		Action:           1,
		Started:          uint32(time.Now().Unix()),
		AssociationToken: utils.RandomString(256),
	}
	_, err = g.repocreatelog_store.Save(ctx, crl)
	if err != nil {
		return nil, err
	}
	//	call ourselfs via external dns name to see if we're are ok with that host
	// hit it via "http" to make sure we hit the right gitserver for the domain and trigger the creation of it
	h := http.HTTP{}
	url := fmt.Sprintf("http://%s/git/self/Create", req.URL.Host)
	h.SetHeader("X-AssociationToken", crl.AssociationToken)
	hb := h.Get(url)
	err = hb.Error()
	if err != nil {
		fmt.Printf("HTTP-Create request failed: %s\n", utils.ErrorString(err))
		return nil, err
	}

	sr.URLs = []*gitpb.SourceRepositoryURL{req.URL}
	return sr, nil
}
