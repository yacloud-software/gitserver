package git2

import (
	"context"
	"flag"
	"fmt"
	"golang.conradwood.net/apis/objectauth"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/utils"
)

var (
	// additional (completed) repos the repobuilder service may read. For example, "skel-go"
	REPOBUILDER_READ     = []uint64{64}
	disable_access_check = flag.Bool("disable_access_check", false, "if true, allow all access")
)

// allows access to the user "by objectauth"
// allows access to repos for repobuilder if repo is not complete yet
func (h *HTTPRequest) hasAccess(ctx context.Context) bool {
	if h.user == nil {
		return false
	}
	if *disable_access_check {
		return true
	}
	// is not complete? access by repobuilder only
	if !h.repo.gitrepo.CreatedComplete {
		if isrepobuilder(ctx) {
			return true
		}
		fmt.Printf("Access to repo %d by repobuilder only\n", h.repo.gitrepo.ID)
		return false
	}
	if !isrepobuilder(ctx) && (h.isWrite() && h.repo.gitrepo.ReadOnly) {
		fmt.Printf("Access to repo %d not granted (repo is readonly)\n", h.repo.gitrepo.ID)
		return false
	}

	err := isDeleted(ctx, h.repo.gitrepo)
	if err != nil {
		fmt.Printf("Access attempt to repo %d (which was deleted): %s\n", h.repo.gitrepo.ID, err)
		return false
	}
	// repobuilder also has some limited (read) access
	if isrepobuilder(ctx) {
		if h.isWrite() {
			fmt.Printf("Repobuilder may not write to repo %d\n", h.repo.gitrepo.ID)
			return false
		}
		for _, r := range REPOBUILDER_READ {
			if r == h.repo.gitrepo.ID {
				return true
			}
		}
		fmt.Printf("Repobuilder may not read repo %d\n", h.repo.gitrepo.ID)
		return false
	}
	// check user access
	objauth := objectauth.GetObjectAuthClient()
	ar, err := objauth.AskObjectAccess(ctx, &objectauth.AuthRequest{
		ObjectType: objectauth.OBJECTTYPE_GitRepository,
		ObjectID:   h.repo.gitrepo.ID,
	})

	if err != nil {
		fmt.Printf("Failed to get object access: %s\n", utils.ErrorString(err))
		return false
	}

	p := ar.Permissions
	if h.isWrite() {
		if p.Write && p.Read && p.View {
			return true
		}
	} else {
		if p.Read && p.View {
			return true
		}
	}
	if auth.IsRoot(ctx) {
		fmt.Printf("Warning - access only allowed because user #%s(%s) is root (repo=%d)\n", h.user.ID, h.user.Email, h.repo.gitrepo.ID)
		return true
	}

	fmt.Printf("Access denied. Repository ID: %d, UserID: %s, Write: %v\n", h.repo.gitrepo.ID, h.user.ID, h.isWrite())
	return false
}

func isrepobuilder(ctx context.Context) bool {
	s := auth.GetService(ctx)
	if s != nil && s.ID == REPO_SERVICE_ID {
		return true
	}
	u := auth.GetUser(ctx)
	if u != nil && u.ID == REPO_SERVICE_ID {
		return true
	}
	return false
}
