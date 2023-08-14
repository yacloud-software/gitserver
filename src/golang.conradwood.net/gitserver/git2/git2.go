package git2

import (
	"flag"
	"fmt"
	"golang.conradwood.net/apis/auth"
	"golang.conradwood.net/apis/common"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/config"
	"golang.conradwood.net/gitserver/crossprocdata"
	"golang.conradwood.net/gitserver/query"
	au "golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/cmdline"
	"golang.conradwood.net/go-easyops/server"
	"golang.conradwood.net/go-easyops/sql"
	"golang.conradwood.net/go-easyops/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	default_user_id         = flag.String("default_user_id", "", "testing only!! if set, and signed_user_is_optional, then requests are processed with this userid")
	signed_user_is_optional = flag.Bool("signed_user_is_optional", false, "normally, the header remote userid requires a corresponding signed user header. setting this to true makes it optional (useful for testing, not for production use)")
	http_port               = flag.Int("git2_http_port", 0, "http-server for git2, tcp port")
	grpc_port               = flag.Int("git2_grpc_port", 0, "grpc-server for git2, tcp port")
	debug                   = flag.Bool("debug_git2", false, "debug git2 server")
	pwd                     string
	psql                    *sql.DB
	/*
		dbhost                  = flag.String("db2host", "localhost", "git2: hostname of the postgres database rdbms")
		dbdb                    = flag.String("db2db", "", "git2: database to use")
		dbuser                  = flag.String("db2user", "root", "git2: username for the database to use")
		dbpw                    = flag.String("db2pw", "pw", "git2: password for the database to use")
	*/
)

func GetDatabase() *sql.DB {
	return psql
}
func Start() error {
	var err error
	if *http_port == 0 {
		return fmt.Errorf("please set -get2_http_port")
	}
	if *grpc_port == 0 {
		return fmt.Errorf("please set -get2_grpc_port")
	}
	pwd, err = os.Getwd()
	if err != nil {
		return err
	}
	if !cmdline.Datacenter() {
		for {
			b := utils.FileExists(pwd + "/.git")
			if b {
				break
			}
			pwd = filepath.Dir(pwd)
			if len(pwd) < 2 {
				return fmt.Errorf("current directory not at a sensible place where we can find scripts\n")
			}
		}
	}
	//	psql, err = sql.OpenWithInfo(*dbhost, *dbdb, *dbuser, *dbpw)
	psql, err = sql.Open()
	if err != nil {
		return err
	}
	gserver := &GIT2{}
	err = gserver.startHTTPServer()
	if err != nil {
		return err
	}

	sd := server.NewHTMLServerDef("gitserver.GIT2")
	sd.Port = *http_port
	server.AddRegistry(sd)
	fmt.Printf("GIT2 http server started on port %d\n", *http_port)

	err = gserver.init()
	if err != nil {
		return err
	}
	sd = server.NewServerDef()
	sd.Port = *grpc_port
	sd.Register = server.Register(func(server *grpc.Server) error {
		gitpb.RegisterGIT2Server(server, gserver)
		return nil
	})
	go func() {
		err := server.ServerStartup(sd)
		utils.Bail("failed to start git2 grpc server", err)
	}()

	return nil
}

func (g *GIT2) startHTTPServer() error {
	if *http_port == 0 {
		return fmt.Errorf("Invalid port %d", *http_port)
	}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *http_port))
	if err != nil {
		return err
	}
	go func() {
		err = http.Serve(ln, g)
		if err != nil {
			fmt.Printf("Failed to serve http over port: %s\n", err)
			os.Exit(10)
		}
	}()
	return nil
}

func (g *GIT2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &HTTPRequest{r: r, w: w, git2: g}
	req.key = utils.RandomString(64)
	crossprocdata.SaveLocalData(req.key, &crossprocdata.LocalData{HTTPRequest: req})
	req.ServeHTTP()
}

type HTTPRequest struct {
	w    http.ResponseWriter
	r    *http.Request
	user *auth.User
	repo *Repo
	gurl *GitURL
	git2 *GIT2
	key  string
}

func (h *HTTPRequest) isWrite() bool {
	if strings.Contains(h.r.URL.RawQuery, "git-receive-pack") {
		return true
	}
	return false
}
func (h *HTTPRequest) isCreate() bool {
	u := h.r.URL.Path
	if strings.HasSuffix(u, `/Create`) {
		return true
	}
	return false
}
func (h *HTTPRequest) isRecreate() bool {
	u := h.r.URL.Path
	if strings.HasSuffix(u, `/Recreate`) {
		return true
	}
	return false
}
func (h *HTTPRequest) isPing() bool {
	u := h.r.URL.Path
	if strings.HasSuffix(u, `/Ping`) {
		return true
	}
	return false
}
func (h *HTTPRequest) pwd() string {
	return pwd
}

func (h *HTTPRequest) ServeHTTP() {
	var err error
	h.gurl = NewGitURL(h.r.URL.Path)

	fmt.Printf("GIT2 ACCESS: %s %s\n", h.r.Host, h.r.URL.Path)

	if h.isPing() {
		query.Ping(h)
		return
	}
	if h.isCreate() {
		h.CreateRepo()
		return
	}
	if h.isRecreate() {
		h.RecreateRepo()
		return
	}
	if h.isRebuild() {
		h.RebuildRepo()
		return
	}

	if strings.Contains(h.r.URL.Path, `/self`) {
		fmt.Printf("Will not allow /self urls (%s)\n", h.r.URL.Path)
		h.ErrorCode(404, "self path not valid")
		return
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
	h.repo, err = RepoFromURL(ctx, h.r.Host, h.gurl)
	if err != nil {
		h.Error(err)
		return
	}
	if h.repo == nil {
		h.ErrorCode(404, "no such repo")
		return
	}
	if !h.repo.ExistsOnDisk() {
		h.ErrorCode(404, "no such repo directory")
		return
	}
	h.Printf("User #%s(%s) Repo: %d (%s) -> %s/%s\n", h.user.ID, h.user.Email, h.repo.gitrepo.ID, h.repo.gitrepo.ArtefactName, h.GitRoot(), h.repo.OnDiskPath())

	if !h.hasAccess(ctx) {
		fmt.Printf("Access denied.\n")
		h.ErrorCode(403, "no such repo directory")
		return
	}
	if !isrepobuilder(ctx) && (h.isWrite() && h.repo.gitrepo.ReadOnly) {
		h.ErrorCode(409, fmt.Sprintf("Write-Access to repo %d not granted (repo is readonly)\n", h.repo.gitrepo.ID))
		return
	}

	REPO_SERVICE_ID := au.GetServiceIDByName("repobuilder.RepoBuilder")
	if h.isWrite() && h.user.ID != REPO_SERVICE_ID {
		if h.repo == nil || h.git2 == nil {
			fmt.Printf("uninitialized request\n")
		} else {
			// update user & commit timestamp
			sr := h.repo.gitrepo
			sr.LastCommit = uint32(time.Now().Unix())
			sr.LastCommitUser = h.user.ID
			sr.UserCommits++
			err = h.git2.repo_store.Update(ctx, sr)
			if err != nil {
				fmt.Printf("failed to update git repo stats: %s\n", utils.ErrorString(err))
			}
		}
	}

	if !h.repo.gitrepo.Forking {
		err = h.fixHooks() // install post-receive etc hooks in git repo
		if err != nil {
			h.Error(err)
			return
		}
	}

	h.InvokeGitCGI(ctx)

}

func (h *HTTPRequest) Error(err error) {
	st := status.Convert(err)
	fmt.Printf("Error: %s\n", utils.ErrorString(err))
	if st.Code() == codes.NotFound {
		h.w.WriteHeader(404)
		return
	}
	h.w.WriteHeader(500)
}

func (h *HTTPRequest) ErrorCode(code int, msg string) {
	h.w.WriteHeader(code)
	h.w.Write([]byte(msg))
}

// true if we have a valid remote_user header and were able to match it to a user by ID
func (h *HTTPRequest) setUser() bool {
	/* also look at http.extraHeader git option.
	[http]
		extraHeader = "REMOTE_USERID: 1"
	*/
	remoteid := h.SingleHeader("remote_userid")
	if remoteid == "" {
		if *signed_user_is_optional {
			h.user = get_current_user()
			if h.user == nil {
				return false
			}
			return true
		}
		fmt.Printf("ERROR: no remote_userid header found\n")
		return false
	}
	suser := h.SingleHeader("X-SIGNEDUSER")
	if suser == "" {
		if !*signed_user_is_optional {
			fmt.Printf("ERROR: no signed user\n")
			return false
		}
		xuser, err := authremote.GetUserByID(authremote.Context(), remoteid)
		if err != nil {
			fmt.Printf("failed to get user: %s\n", utils.ErrorString(err))
			return false
		}
		h.user = xuser
		return true
	}
	user := &auth.User{}
	err := utils.Unmarshal(suser, user)
	if err != nil {
		fmt.Printf("ERROR: signed user cannot be unmarshaled (%s)", utils.ErrorString(err))
		return false
	}

	if user == nil {
		fmt.Printf("No user for id %s\n", remoteid)
		return false
	}
	fmt.Printf("From remoteid: \"%s\", from signeduser: \"%s\"\n", remoteid, user.ID)
	h.user = user
	return true
}

// if all goes well send this response and repeat the header "X-PleaseRepeatBack"
func (g *HTTPRequest) Respond(txt string) {
	s := g.SingleHeader(REPEAT_BACK_HEADER)
	r := txt + " " + s
	g.Write([]byte(r))
}
func (g *HTTPRequest) Write(b []byte) {
	g.w.Write(b)
}

// get a single header from the request
func (g *HTTPRequest) SingleHeader(name string) string {
	h := g.r.Header[name]
	if h == nil {
		for k, v := range g.r.Header {
			if strings.ToLower(k) == strings.ToLower(name) {
				h = v
				break
			}
		}
	}
	if len(h) > 0 {
		return h[0]
	}
	return ""
}
func (h *HTTPRequest) Username() string {
	u := h.user
	if u == nil {
		return "ANONYMOUS"
	}
	return fmt.Sprintf("%s(%s)", u.Email, u.ID)
}

func (h *HTTPRequest) Printf(format string, args ...interface{}) {
	prefix := "[" + h.Username() + "] "
	fmt.Printf(prefix+format, args...)

}
func (h *HTTPRequest) GitRoot() string {
	return *config.Gitroot
}
func (h *HTTPRequest) GetScriptDir() string {
	sc := h.pwd() + "/scripts"
	return sc
}

// get the currently running user (of this process)
func get_current_user() *auth.User {
	if *default_user_id != "" {
		ctx, err := authremote.ContextForUserID(*default_user_id)
		if err != nil {
			fmt.Printf("failed to get context for user \"%s\": %s\n", *default_user_id, utils.ErrorString(err))
			return nil
		}
		u := au.GetUser(ctx)
		if u == nil {
			fmt.Printf("failed to get user from context\n")
			return nil
		}
		fmt.Printf("Running as user: \"%s\"\n", au.Description(u))
		return u

	}
	ctx := authremote.Context()
	u, err := authremote.GetAuthManagerClient().WhoAmI(ctx, &common.Void{})
	if err != nil {
		fmt.Printf("failed to whoami(): %s\n", err)
		return nil
	}
	if u == nil {
		fmt.Printf("No remote_userid, and context has no user either\n")
		return nil
	}
	fmt.Printf("User: %v\n", au.Description(u))
	return u

}
