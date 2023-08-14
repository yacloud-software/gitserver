package git2

import (
	"context"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/builder"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
	"strconv"
	"strings"
)

func (h *HTTPRequest) isRebuild() bool {
	u := h.r.URL.Path
	if strings.Contains(u, `/Rebuild/`) {
		return true
	}
	return false
}
func (h *HTTPRequest) RebuildRepo() {
	rebstr := "/Rebuild/"
	idx := strings.Index(h.r.URL.Path, rebstr)
	if idx == -1 {
		fmt.Printf("Missing \"%s\" in Path (%s)\n", rebstr, h.r.URL.Path)
		return
	}
	ids := h.r.URL.Path[idx+len(rebstr):]
	buildid, err := strconv.ParseUint(ids, 10, 64)
	if err != nil {
		fmt.Printf("Invalid id: %s\n", err)
		return
	}
	rr := &gitpb.RebuildRequest{
		ID:                  buildid,
		ExcludeBuildScripts: []string{"DIST"},
	}
	gotuser := h.setUser()
	if !gotuser {
		h.ErrorCode(401, "authentication required")
		return
	}

	ctx, err := authremote.ContextForUser(h.user)
	if err != nil {
		h.Error(err)
		return
	}

	h.w.Header().Set("X-gitserver-build", "true")

	build, err := db.DefaultDBBuild().ByID(ctx, buildid)
	if err != nil {
		fmt.Printf("Failed to get build: %s\n", err)
		h.ErrorCode(404, fmt.Sprintf("repo %d not found", buildid))
		return
	}
	sr, err := db.DefaultDBSourceRepository().ByID(ctx, build.RepositoryID)
	if err != nil {
		fmt.Printf("Failed to get repo for build: %s\n", err)
		h.ErrorCode(404, fmt.Sprintf("repo %d not found", buildid))
		return
	}
	err = wantRepoAccess(ctx, sr, false) //readaccess?
	if err != nil {
		fmt.Printf("access denied:% s", err)
		h.ErrorCode(403, "access denied")
		return
	}
	h.w.Header().Set("X-gitserver-repo", fmt.Sprintf("%d", sr.ID))
	h.w.Header().Set("X-gitserver-artefact", sr.ArtefactName)
	fmt.Printf("Rebuilding (via url) Build %d in Repository %d (%s)\n", buildid, sr.ID, sr.ArtefactName)
	err = rebuild(rr, NewHTTPWriter(h, ctx))
	if err != nil {
		fmt.Printf("Rebuild encountered error: %s\n", err)
		h.Write([]byte(fmt.Sprintf("Rebuild encountered error: %s\n", err)))
	}

}
func (g *GIT2) Rebuild(req *gitpb.RebuildRequest, srv gitpb.GIT2_RebuildServer) error {
	return rebuild(req, srv)
}

type rebuild_server interface {
	Context() context.Context
	Send(hr *gitpb.HookResponse) error
}

func rebuild(req *gitpb.RebuildRequest, srv rebuild_server) error {
	ctx := srv.Context()
	user := auth.GetUser(ctx)
	if user == nil {
		return errors.Unauthenticated(ctx, "please log in")
	}
	buildid := req.ID
	if buildid == 0 {
		return errors.NotFound(ctx, "cannot rebuild build #0")
	}
	fmt.Printf("Request to rebuild #%d by user %s\n", req.ID, auth.Description(user))
	build, err := db.DefaultDBBuild().ByID(ctx, buildid)
	if err != nil {
		return err
	}
	sr, err := db.DefaultDBSourceRepository().ByID(ctx, build.RepositoryID)
	if err != nil {
		return err
	}
	err = wantRepoAccess(ctx, sr, false) //readaccess?
	if err != nil {
		return err
	}

	w := &HookResponseWriter{srv: srv}
	rt := &RebuildTrigger{
		ctx:             ctx,
		userid:          user.ID,
		repoid:          sr.ID,
		newrev:          build.CommitHash,
		branch:          build.Branch,
		excludedscripts: req.ExcludeBuildScripts,
	}
	fmt.Printf("Rebuilding Build #%d for user \"%s\"\n", buildid, auth.UserIDString(user))
	lr, err := builder.RunExternalBuilder(ctx, rt, build.ID, w)
	if lr != nil {
		fmt.Printf("Last response: %s\n", lr.ResultMessage)
		fmt.Printf("Success: %v\n", lr.Success)
	}
	if err != nil {
		return err
	}
	if !build.Success {
		build.Success = true
		err = db.DefaultDBBuild().Update(ctx, build)
		if err != nil {
			fmt.Printf("Warning - failed to update last build status in database from failed to success\n")
		}
	}
	return nil
}

type RebuildTrigger struct {
	ctx             context.Context
	userid          string
	repoid          uint64
	newrev          string
	branch          string
	excludedscripts []string
}

func (r *RebuildTrigger) ExcludeBuildScripts() []string {
	return r.excludedscripts
}
func (r *RebuildTrigger) RepositoryID() uint64 {
	return r.repoid
}
func (r *RebuildTrigger) GetContext() (context.Context, error) {
	return r.ctx, nil
}
func (r *RebuildTrigger) NewRev() string {
	return r.newrev
}
func (r *RebuildTrigger) Branch() string {
	return r.branch
}
func (r *RebuildTrigger) UserID() string {
	return r.userid
}

type hookresponsereceiver interface {
	Send(*gitpb.HookResponse) error
}
type HookResponseWriter struct {
	srv hookresponsereceiver
}

func (h *HookResponseWriter) Write(buf []byte) (int, error) {
	hr := &gitpb.HookResponse{}
	hr.Output = string(buf)
	err := h.srv.Send(hr)
	if err != nil {
		return 0, err
	}
	return len(buf), nil
}
