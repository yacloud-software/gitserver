package git2

import (
	"context"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/builder"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/errors"
)

func (g *GIT2) Rebuild(req *gitpb.ByIDRequest, srv gitpb.GIT2_RebuildServer) error {
	ctx := srv.Context()
	user := auth.GetUser(ctx)
	if user == nil {
		return errors.Unauthenticated(ctx, "please log in")
	}
	buildid := req.ID
	if buildid == 0 {
		return errors.NotFound(ctx, "cannot rebuild build #0")
	}

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
		ctx:    ctx,
		userid: user.ID,
		repoid: sr.ID,
		newrev: build.CommitHash,
		branch: build.Branch,
	}
	lr, err := builder.RunExternalBuilder(ctx, rt, build.ID, w)
	if lr != nil {
		fmt.Printf("Last response: %s\n", lr.ResultMessage)
		fmt.Printf("Success: %v\n", lr.Success)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Rebuilding Build #%d for user \"%s\"\n", buildid, auth.UserIDString(user))
	return nil
}

type RebuildTrigger struct {
	ctx    context.Context
	userid string
	repoid uint64
	newrev string
	branch string
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