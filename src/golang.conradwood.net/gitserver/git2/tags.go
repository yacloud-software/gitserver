package git2

import (
	"context"

	"golang.conradwood.net/apis/common"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/errors"
)

func (g *GIT2) AttachNTagToBuild(ctx context.Context, req *gitpb.AttachNTagRequest) (*common.Void, error) {
	if req.BuildID == 0 {
		return nil, errors.InvalidArgs(ctx, "missing buildid", "missing buildid")
	}
	ntres, err := g.GetBuildWithNTag(ctx, &gitpb.GetNTagRequest{RepositoryID: req.RepositoryID, Tag: req.Tag})
	if err != nil {
		return nil, err
	}
	build, err := db.DefaultDBBuild().ByID(ctx, req.BuildID)
	if err != nil {
		return nil, err
	}
	if build.RepositoryID != req.RepositoryID {
		return nil, errors.InvalidArgs(ctx, "mismatch build/repositoryid", "build #%d belongs to repository %d instead of %d", req.BuildID, build.RepositoryID, req.RepositoryID)
	}

	if ntres.NTag == nil {
		nt := &gitpb.NRepoTagID{
			RepositoryID: req.RepositoryID,
			Tag:          req.Tag,
			BuildID:      req.BuildID,
		}
		_, err = db.DefaultDBNRepoTagID().Save(ctx, nt)
	} else {
		ntres.NTag.BuildID = req.BuildID
		err = db.DefaultDBNRepoTagID().Update(ctx, ntres.NTag)
	}
	if err != nil {
		return nil, err
	}
	return &common.Void{}, nil
}
func (g *GIT2) GetBuildWithNTag(ctx context.Context, req *gitpb.GetNTagRequest) (*gitpb.NTagResponse, error) {
	if req.RepositoryID == 0 {
		return nil, errors.InvalidArgs(ctx, "missing repositoryid", "missing repositoryid")
	}
	err := wants_access_to_build(ctx, req.RepositoryID) // despite the name, input *is* a repositoryid
	if err != nil {
		return nil, err
	}
	q := db.DefaultDBNRepoTagID().NewQuery()
	q.AddEqual("repositoryid", req.RepositoryID)
	q.AddEqual("tag", req.Tag)
	builds, err := db.DefaultDBNRepoTagID().ByDBQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	res := &gitpb.NTagResponse{}
	if len(builds) != 0 {
		res.NTag = builds[0]
	}
	return res, nil
}
