package git2

import (
	"context"
	"fmt"
	pb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/config"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/utils"
	"log"
	"net/http"
	"net/http/cgi"
	"net/url"
	"strings"
)

const (
	gitbin = "/usr/lib/git-core/git-http-backend"
)

func (h *HTTPRequest) InvokeGitCGI(ctx context.Context) {
	if !utils.FileExists(gitbin) {
		panic(fmt.Sprintf("Git http backend missing (%s)", gitbin))
	}
	var err error
	// create a new 'fake' request for the cgihandler
	// with the path adjusted to match the filesystem (rather than URL)
	// the original path might be something like /git/yacloud.eu/test.git/info/refs?...
	// the rewritten path will be /byid/1.git/info/refs?...
	//
	// if a repo is currently 'forking' then:
	// a) write access is denied
	// b) read access is served from forkedfrom repo (source repo)
	newreq := *h.r

	on := h.r.URL.String()
	idx := strings.Index(on, ".git")
	if idx == -1 {
		h.Error(fmt.Errorf("missing .git in url"))
		return
	}
	on = on[idx+4:]
	p := h.repo.OnDiskPath() + on
	if h.repo.gitrepo.Forking {
		fmt.Printf("Warning - Access to repo %d, but it is still forking\n", h.repo.gitrepo.ID)
		h.repo.forkedRepo, err = db.NewDBSourceRepository(psql).ByID(ctx, h.repo.gitrepo.ForkedFrom)
		if err != nil {
			h.Error(err)
			return
		}
		copyrepo(h.repo.forkedRepo, h.repo.gitrepo)
		if h.isWrite() {
			fmt.Printf("WRITE Access denied to repo %d BECAUSE it is still forking\n", h.repo.gitrepo.ID)
			h.Error(fmt.Errorf("Repo %d not ready yet (cloning)", h.repo.gitrepo.ID))
			return
		}
		p = strings.Trim(h.repo.forkedRepo.FilePath, "/") + on
		fmt.Printf("Serving read-request repo from to-be-forked-from repo filepath: \"%s\"\n", p)
	}

	newurl := fmt.Sprintf("https://%s/%s", h.r.Host, p)
	if *debug {
		h.Printf("Original URL: \"%s\"\n", h.r.URL.String())
		h.Printf("Raw Fake URL: \"%s\"\n", newurl)
	}
	newreq.URL, err = url.ParseRequestURI(newurl)
	if err != nil {
		fmt.Printf("Failed to parse fakeurl \"%s\": %s\n", newurl, err)
		h.Error(err)
		return
	}
	if *debug {
		h.Printf("    Fake URL: \"%s\"\n", newreq.URL.String())
	}
	//	newreq.Method = h.r.Method
	// set up the environment for the cgi

	if h.isWrite() {
		h.informAdminsAboutCommit()
	}
	// the gitinfo which is passed through back to us through the hooks
	gi := &pb.GitInfo{Version: 2, RepositoryID: h.repo.gitrepo.ID, UserID: h.user.ID, User: h.user, URL: h.r.URL.String()}
	gp, err := utils.Marshal(gi)
	if err != nil {
		h.Error(err)
		return
	}

	// serialised context to use for hooks/shellscripts
	ncb, err := auth.SerialiseContextToString(ctx)
	if err != nil {
		h.Error(err)
		return
	}
	ncs := fmt.Sprintf("GE_CTX=%s", ncb)

	s := h.GitRoot()
	gitroot := &s
	ch := cgi.Handler{
		Path: gitbin,
		Env: func() (env []string) {
			env = append(env, ncs)
			env = append(env, fmt.Sprintf("GITSERVER_DIR=%s", h.pwd()))
			env = append(env, fmt.Sprintf("GITSERVER_TCP_PORT=%d", *config.Gitport))
			env = append(env, fmt.Sprintf("GITINFO=%s", gp))
			env = append(env, fmt.Sprintf("REMOTE_USER=%s", h.user.ID))
			env = append(env, fmt.Sprintf("REPOSITORY_ID=%d", h.repo.gitrepo.ID))
			env = append(env, "GIT_HTTP_EXPORT_ALL=")
			env = append(env, "REMOTE_ADDR=local") // backend sets GIT_COMMITTER_EMAIL to ${REMOTE_USER}@http.${REMOTE_ADDR},
			env = append(env, "GIT_HTTP_MAX_REQUEST_BUFFER=100M")
			env = append(env, fmt.Sprintf("GIT_PROJECT_ROOT=%s", *gitroot))
			return
		}(),
	}
	if *debug {
		fmt.Printf("debug - Invoking CGI...\n")
		ch.Logger = log.New(&cgilogger{}, "[cgi]", 0)
	}
	crw := &cgiResponseWriter{h: h}
	ch.ServeHTTP(crw, &newreq)

}

type cgilogger struct {
}

func (c *cgilogger) Write(buf []byte) (int, error) {
	fmt.Printf("%s\n", string(buf))
	return len(buf), nil
}

type cgiResponseWriter struct {
	h *HTTPRequest
}

func (c *cgiResponseWriter) Header() http.Header {
	return c.h.w.Header()
}
func (c *cgiResponseWriter) Write(buf []byte) (int, error) {
	if *debug {
		fmt.Printf("cgi said: %s\n", string(buf))
	}
	n, err := c.h.w.Write(buf)
	if err != nil {
		return n, err
	}
	flusher, isFlusher := c.h.w.(http.Flusher) // some http responsewriters implement "flusher" interface
	if isFlusher {
		//		fmt.Printf("flushing cgi...\n")
		flusher.Flush()
	}
	return n, err

}
func (c *cgiResponseWriter) WriteHeader(statusCode int) {
	c.h.w.WriteHeader(statusCode)
}
