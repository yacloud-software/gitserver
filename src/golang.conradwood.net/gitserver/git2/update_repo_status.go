package git2

import (
	"context"
	gitpb "golang.conradwood.net/apis/gitserver"
	"strings"
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

	return sr, nil
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

func (g *GIT2) GitRepoUpdate(ctx context.Context, req *gitpb.RepoUpdateRequest) (*gitpb.SourceRepository, error) {
	if req.Original == nil || req.Original.ID == 0 {
		return nil, errors.InvalidArgs(ctx, "missing repository id", "missing repository id")
	}
	ri, err := g.RepoByID(ctx, &gitpb.ByIDRequest{ID: req.Original.ID})
	if err != nil {
		return nil, err
	}

	if req.AddURLHost != "" && req.AddURLPath != "" {
		cr, err := g.CheckGitServer(ctx, &gitpb.CheckGitRequest{Host: req.AddURLHost})
		if err != nil {
			return nil, err
		}
		if !cr.Success {
			return nil, errors.InvalidArgs(ctx, "no gitserver found at host \"%s\"", req.AddURLHost)
		}
		path := req.AddURLPath
		if strings.HasSuffix(path, ".git") {
			return nil, errors.InvalidArgs(ctx, "paths must end with .git", "paths must end with .git")
		}
		// TODO: append .git if necessary, check for dupes etc
		ur := &gitpb.SourceRepositoryURL{
			V2RepositoryID: ri.ID,
			Host:           req.AddURLHost,
			Path:           path,
		}
		_, err = g.url_store.Save(ctx, ur)
		if err != nil {
			return nil, err
		}
		ri.URLs = append(ri.URLs, ur)
	}

	save := false
	if req.Description != "" {
		save = true
		ri.Description = req.Description
	}

	if save {
		err = g.repo_store.Update(ctx, ri)
		if err != nil {
			return nil, err
		}
	}

	return ri, nil

}

