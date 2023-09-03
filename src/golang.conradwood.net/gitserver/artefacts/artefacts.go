package artefacts

import (
	"context"
	"fmt"
	af "golang.conradwood.net/apis/artefact"
	br "golang.conradwood.net/apis/buildrepo"
	"golang.conradwood.net/apis/common"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/authremote"
)

func CreateIfRequired(ctx context.Context, repo *gitpb.SourceRepository) (*af.CreateArtefactResponse, error) {
	car := &af.CreateArtefactRequest{
		OrganisationID: "gitserver_not_supported",
		ArtefactName:   repo.ArtefactName,
	}
	// if no urls, load them from database
	if len(repo.URLs) == 0 {
		urls, err := db.DefaultDBSourceRepositoryURL().ByV2RepositoryID(ctx, repo.ID)
		if err != nil {
			return nil, err
		}
		repo.URLs = urls
	}
	if len(repo.URLs) == 0 {
		return nil, fmt.Errorf("Repository %d (%s) has no urls\n", repo.ID, repo.ArtefactName)
	}
	u := repo.URLs[0]
	car.GitURL = fmt.Sprintf("https://%s/%s", u.Host, u.Path)

	// get our default build domain  (from buildrepo)
	bi, err := br.GetBuildRepoManagerClient().GetManagerInfo(ctx, &common.Void{})
	if err != nil {
		return nil, err
	}
	car.BuildRepoDomain = bi.Domain
	afm, err := af.GetArtefactServiceClient().CreateArtefactIfRequired(ctx, car)
	if err != nil {
		return nil, err
	}
	return afm, nil
}
func RepositoryIDToArtefactID(id uint64) (uint64, error) {
	req := &af.ID{ID: id}
	ctx := authremote.Context()
	afid, err := af.GetArtefactServiceClient().GetArtefactForRepo(ctx, req)
	if err != nil {
		fmt.Printf("Failed to resolve repositoryid %d to artefact: %s\n", id, err)
		return 0, err
	}
	if afid.ID == 0 {
		fmt.Printf("artefact server resolved repositoryid %d to artefactid 0!\n", id)
	}
	return afid.ID, nil
}
