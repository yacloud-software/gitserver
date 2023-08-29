package artefacts

import (
	af "golang.conradwood.net/apis/artefact"
	"golang.conradwood.net/go-easyops/authremote"
)

func RepositoryIDToArtefactID(id uint64) (uint64, error) {
	req := &af.ID{ID: id}
	ctx := authremote.Context()
	afid, err := af.GetArtefactServiceClient().GetArtefactForRepo(ctx, req)
	if err != nil {
		return 0, err
	}
	return afid.ID, nil
}
