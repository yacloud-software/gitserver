package git2

import (
	"context"
	gitpb "golang.conradwood.net/apis/gitserver"
	//	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/errors"
)

func (g *GIT2) UpdateRepoStatus(ctx context.Context, req *gitpb.UpdateRepoStatusRequest) (*gitpb.SourceRepository, error) {
	/*
		u := auth.GetUser(ctx)
		if u == nil {
			return nil, errors.Unauthenticated(ctx, "please authenticate")
		}
	*/
	if !isrepobuilder(ctx) {
		return nil, errors.AccessDenied(ctx, "access only for repobuilder")
	}
	if req.RepoID == 0 {
		return nil, errors.InvalidArgs(ctx, "missing repository id", "missing repository id")
	}
	bir := &gitpb.ByIDRequest{ID: req.RepoID}
	sr, err := g.RepoByID(ctx, bir)
	if err != nil {
		return nil, err
	}
	nb, set := decode_repostate(req.ReadOnly)
	if set {
		sr.ReadOnly = nb
	}
	nb, set = decode_repostate(req.RunHooks)
	if set {
		sr.RunPostReceive = nb
		sr.RunPreReceive = nb
	}
	err = g.repo_store.Update(ctx, sr)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
func decode_repostate(i gitpb.NewRepoState) (bool, bool) {
	if i == gitpb.NewRepoState_SET_TRUE {
		return true, true
	}
	if i == gitpb.NewRepoState_SET_FALSE {
		return false, true
	}
	return false, false
}
