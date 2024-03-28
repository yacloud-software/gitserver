package git2

import (
	"context"
	"flag"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/apis/objectauth"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/utils"
)

var (
	PRIVILEGED_SERVICES = []string{
		auth.GetServiceIDByName("firmwaretracker.FirmwareTracker"),
		auth.GetServiceIDByName("espota.ESPOtaService"),
		auth.GetServiceIDByName("autodeployer.AutoDeployer"),
	}
	// additional (completed) repos the repobuilder service may read. For example, "skel-go"
	REPOBUILDER_READ     = []uint64{64} // a gitrepoid (e.g. skel-go)
	disable_access_check = flag.Bool("disable_access_check", false, "if true, allow all access")
)

func is_privileged_service(ctx context.Context) bool {
	svc := auth.GetService(ctx)
	if svc == nil {
		return false
	}
	for _, sa := range PRIVILEGED_SERVICES {
		if svc.ID == sa {
			return true
		}
	}
	if HasServiceAnyAccess(ctx, objectauth.OBJECTTYPE_GitRepository) {
		return true
	}

	return false
}

// nil if ok
func wantRepoAccess(ctx context.Context, repo *gitpb.SourceRepository, writereq bool) error {
	if is_privileged_service(ctx) {
		return nil
	}
	objauth := objectauth.GetObjectAuthClient()
	ar, err := objauth.AskObjectAccess(ctx, &objectauth.AuthRequest{
		ObjectType: objectauth.OBJECTTYPE_GitRepository,
		ObjectID:   repo.ID,
	})

	if err != nil {
		fmt.Printf("Failed to get object access: %s\n", utils.ErrorString(err))
		return err
	}

	p := ar.Permissions
	if writereq {
		if p.Write && p.Read && p.View {
			return nil
		}
	} else {
		if p.Read && p.View {
			return nil
		}
	}

	return errors.AccessDenied(ctx, "access to repo %d (%s) denied", repo.ID, repo.ArtefactName)
}

// allows access to the user "by objectauth"
// allows access to repos for repobuilder if repo is not complete yet
func (h *HTTPRequest) hasAccess(ctx context.Context) bool {
	if is_privileged_service(ctx) {
		return true
	}
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

	err := isDeleted(ctx, h.repo.gitrepo)
	if err != nil {
		fmt.Printf("Access attempt to repo %d (which was deleted): %s\n", h.repo.gitrepo.ID, err)
		return false
	}
	// repobuilder also has some limited (read) access
	if isrepobuilder(ctx) {
		if h.isWrite() {
			fmt.Printf("Repobuilder may write to any repo %d\n", h.repo.gitrepo.ID)
			return true
		}
		for _, r := range REPOBUILDER_READ {
			if r == h.repo.gitrepo.ID {
				return true
			}
		}
		//		fmt.Printf("Repobuilder may not read repo %d\n", h.repo.gitrepo.ID)
		//		return false
		fmt.Printf("Repobuilder may read any repo %d\n", h.repo.gitrepo.ID)
		return true
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
	REPO_SERVICE_ID := auth.GetServiceIDByName("repobuilder.RepoBuilder")
	if s != nil && s.ID == REPO_SERVICE_ID {
		return true
	}
	u := auth.GetUser(ctx)
	if u != nil && u.ID == REPO_SERVICE_ID {
		return true
	}
	return false
}

// despite the name - input parameter is a repositoryid
func wants_access_to_build(ctx context.Context, repoid uint64) error {
	if HasServiceAnyAccess(ctx, objectauth.OBJECTTYPE_GitRepository) {
		return nil
	}
	check_user := true
	serr := errors.NeedServiceOrRoot(ctx, append([]string{WEB_SERVICE_ID, GOTOOLS_SERVICE_ID}, PRIVILEGED_SERVICES...))
	if serr == nil {
		check_user = false
	}

	if check_user {
		ot := &objectauth.AuthRequest{ObjectType: objectauth.OBJECTTYPE_GitRepository, ObjectID: repoid}
		ol, err := objectauth.GetObjectAuthClient().AskObjectAccess(ctx, ot)
		if err != nil {
			return err
		}
		if ol == nil {
			fmt.Printf("for a weird reason we did not get a permissions object but no error either")
			return fmt.Errorf("permission error")
		}
		if ol.Permissions == nil {
			return errors.AccessDenied(ctx, "access denied to git repository %d (no permission)", repoid)
		}
		if !ol.Permissions.Read {
			return errors.AccessDenied(ctx, "access denied to git repository %d (no read permission)", repoid)
		}
	}
	return nil
}
