package artefacts

import (
	"fmt"
	af "golang.conradwood.net/apis/artefact"
	"golang.conradwood.net/go-easyops/authremote"
)

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
