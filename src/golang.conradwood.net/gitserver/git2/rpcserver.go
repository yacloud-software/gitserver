package git2

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.conradwood.net/apis/common"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/apis/objectauth"
	oa "golang.conradwood.net/apis/objectauth"
	"golang.conradwood.net/gitserver/checks"
	"golang.conradwood.net/gitserver/crossprocdata"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/gitserver/query"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/cache"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/http"
	"golang.conradwood.net/go-easyops/utils"
)

var (
	WEB_SERVICE_ID     = auth.GetServiceIDByName("scweb.SCWebService")
	GOTOOLS_SERVICE_ID = auth.GetServiceIDByName("gotools.GoTools")
	// these repos may be 'reset' back to 'bare' at any time
	RESETTABLE_REPOS = []uint64{216}
	urlCache         = cache.NewResolvingCache("giturlcache", time.Duration(15)*time.Minute, 1000)
	debug_githook    = flag.Bool("debug_githook", false, "debug print server side githook stuff")
)

type GIT2 struct {
	url_store           *db.DBSourceRepositoryURL
	repo_store          *db.DBSourceRepository
	repocreatelog_store *db.DBCreateRepoLog
	build_store         *db.DBBuild
}

func (g *GIT2) init() error {
	db.DefaultDBNRepoTagID()
	g.url_store = db.DefaultDBSourceRepositoryURL()
	g.repo_store = db.DefaultDBSourceRepository()
	g.repocreatelog_store = db.DefaultDBCreateRepoLog()
	g.build_store = db.DefaultDBBuild()
	return nil
}
func (g *GIT2) SetDenyMessage(ctx context.Context, req *gitpb.DenyMessageRequest) (*common.Void, error) {
	repo, err := g.RepoByID(ctx, &gitpb.ByIDRequest{ID: req.RepositoryID})
	if err != nil {
		return nil, err
	}
	repo.DenyMessage = req.DenyMessage
	err = g.repo_store.Update(ctx, repo)
	if err != nil {
		return nil, err
	}
	return &common.Void{}, nil
}
func (g *GIT2) RepoByURL(ctx context.Context, req *gitpb.ByURLRequest) (*gitpb.SourceRepository, error) {
	sr, err := g.FindRepoByURL(ctx, req)
	if err != nil {
		return nil, err
	}
	if !sr.Found {
		return nil, errors.NotFound(ctx, "repository not found")
	}
	return sr.Repository, nil
}
func (g *GIT2) RepoByID(ctx context.Context, req *gitpb.ByIDRequest) (*gitpb.SourceRepository, error) {
	if !isrepobuilder(ctx) && !*disable_access_check && !is_privileged_service(ctx) {
		//fmt.Printf("Getting repo by id %d\n", req.ID)
		ot := &oa.AuthRequest{ObjectType: oa.OBJECTTYPE_GitRepository, ObjectID: req.ID}
		ol, err := oa.GetObjectAuthClient().AskObjectAccess(ctx, ot)
		if err != nil {
			return nil, err
		}
		if ol == nil {
			fmt.Printf("for a weird reason we did not get a permissions object but no error either")
			return nil, errors.Errorf("permission error")
		}
		if ol.Permissions == nil {
			return nil, errors.AccessDenied(ctx, "access denied to git repository %d (no permission)", req.ID)
		}
		if !auth.IsRoot(ctx) {
			if !ol.Permissions.View && !ol.Permissions.Read && !ol.Permissions.Execute {
				return nil, errors.AccessDenied(ctx, "access denied to git repository %d", req.ID)
			}
		}
	}
	repo, err := g.repo_store.ByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	err = isDeleted(ctx, repo)
	if err != nil {
		return nil, err
	}

	urls, err := g.url_store.ByV2RepositoryID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	repo.URLs = urls
	return repo, nil
}

func (g *GIT2) GetRepos(ctx context.Context, req *common.Void) (*gitpb.SourceRepositoryList, error) {
	return g.GetReposWithFilter(ctx, &gitpb.RepoFilter{})
}
func (g *GIT2) GetReposWithFilter(ctx context.Context, req *gitpb.RepoFilter) (*gitpb.SourceRepositoryList, error) {
	repos, err := g.repo_store.All(ctx)
	if err != nil {
		return nil, err
	}
	res := &gitpb.SourceRepositoryList{}

	// disable "normal" access check and return all repos matching the filter
	use_all_objects := false

	if HasServiceReadAccess(ctx, oa.OBJECTTYPE_GitRepository) {
		use_all_objects = true
	}

	ol := &oa.ObjectIDList{}
	if !use_all_objects {
		u := auth.GetUser(ctx)
		if u == nil {
			return nil, errors.Unauthenticated(ctx, "no user: no repos")
		}
		//	fmt.Printf("GetRepos() - got %d repos\n", len(repos))
		ot := &oa.ObjectType{ObjectType: oa.OBJECTTYPE_GitRepository}
		ol, err = oa.GetObjectAuthClient().AvailableObjects(ctx, ot)
		if err != nil {
			return nil, err
		}
	}
	for _, sr := range repos {
		if sr.Deleted {
			continue
		}
		b, err := filterMatch(ctx, sr, req)
		if err != nil {
			return nil, err
		}
		if !b {
			continue
		}
		include := false
		if use_all_objects {
			include = true
		} else {
			for _, id := range ol.ObjectIDs {
				if id == sr.ID {
					include = true
					break
				}
			}
		}
		if include || *disable_access_check {
			res.Repos = append(res.Repos, sr)

			uao, err := urlCache.Retrieve(fmt.Sprintf("%d", sr.ID), func(k string) (interface{}, error) {
				return g.url_store.ByV2RepositoryID(ctx, sr.ID)
			})
			if err != nil {
				urlCache.Evict(fmt.Sprintf("%d", sr.ID))
				return nil, err
			}
			sr.URLs = uao.([]*gitpb.SourceRepositoryURL)
		}
	}
	//fmt.Printf("GetRepos() - returning %d repos to user %s (%s)\n", len(res.Repos), auth.Description(u), u.ID)
	sort.Slice(res.Repos, func(i, j int) bool {
		return res.Repos[i].ID < res.Repos[j].ID
	})
	return res, nil
}

/*
* this creates the database entries, but not the file structure (it does not involve calling the git binary
* instead it makes a call via https to the domain&path specified with a special header.
* this ensures we're creating the files on a gitserver that is triggered by the url
* the call is authenticated via a temporary associationtoken
 */
func (g *GIT2) CreateRepo(ctx context.Context, req *gitpb.CreateRepoRequest) (*gitpb.SourceRepository, error) {
	if req.URL == nil {
		return nil, errors.InvalidArgs(ctx, "missing url", "missing url")
	}
	if req.ArtefactName == "" {
		return nil, errors.InvalidArgs(ctx, "missing artefactname", "missing artefactname")
	}
	if req.URL.Path == "" {
		return nil, errors.InvalidArgs(ctx, "missing path", "missing path")
	}
	if strings.HasPrefix(req.URL.Path, "/git/") {
		return nil, errors.InvalidArgs(ctx, "url path must not not start with '/git/' ", "url path must not not start with '/git/' ")
	}
	if req.URL.Host == "" {
		return nil, errors.InvalidArgs(ctx, "missing host", "missing host")
	}
	if req.Description == "" {
		return nil, errors.InvalidArgs(ctx, "description required", "a description is required for each repository")
	}
	if !strings.HasSuffix(req.URL.Path, ".git") {
		req.URL.Path = req.URL.Path + ".git"
	}
	req.URL.Host = strings.ToLower(req.URL.Host)
	err := checkValidHost(ctx, req.URL.Host)
	if err != nil {
		return nil, err
	}
	// check if url already exists in database:
	paths, err := g.url_store.ByPath(ctx, req.URL.Path)
	if err != nil {
		return nil, err
	}
	for _, p := range paths {
		if p.Host == req.URL.Host {
			return g.create_again(ctx, p, req)
		}
	}
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "need user to create repo")
	}
	sr := req2repo(req) // create repo from request
	// create in database
	_, err = g.repo_store.Save(ctx, sr)
	if err != nil {
		return nil, err
	}
	sr.FilePath = fmt.Sprintf("byid/%d", sr.ID)
	err = g.repo_store.Update(ctx, sr)
	if err != nil {
		return nil, err
	}
	req.URL.V2RepositoryID = sr.ID
	_, err = g.url_store.Save(ctx, req.URL)
	if err != nil {
		return nil, err
	}

	// add the create repo log
	ctxs, err := auth.SerialiseContextToString(ctx)
	if err != nil {
		return nil, err
	}
	crl := &gitpb.CreateRepoLog{
		RepositoryID:     sr.ID,
		UserID:           u.ID,
		Context:          ctxs,
		Action:           1,
		Started:          uint32(time.Now().Unix()),
		AssociationToken: utils.RandomString(256),
	}
	_, err = g.repocreatelog_store.Save(ctx, crl)
	if err != nil {
		return nil, err
	}
	// hit it via "http" to make sure we hit the right gitserver for the domain and trigger the creation of it
	h := http.HTTP{}
	url := fmt.Sprintf("https://%s/git/self/Create", req.URL.Host)
	h.SetHeader("X-AssociationToken", crl.AssociationToken)
	hb := h.Get(url)
	err = hb.Error()
	if err != nil {
		fmt.Printf("HTTP-Create request failed: %s\n", utils.ErrorString(err))
		return nil, err
	}

	userids := []string{crl.UserID}
	if crl.UserID == "1" {
		userids = append(userids, "7") // add singingcat
	}
	if crl.UserID == "7" {
		userids = append(userids, "1") // add cnw
	}
	for _, userid := range userids {
		fmt.Printf("Granting access to user id %s\n", userid)
		oreq := &objectauth.GrantUserRequest{
			ObjectType: objectauth.OBJECTTYPE_GitRepository,
			ObjectID:   crl.RepositoryID,
			UserID:     userid,
			Read:       true,
			Write:      true,
			Execute:    true,
			View:       true,
		}
		_, err = objectauth.GetObjectAuthClient().GrantToUser(ctx, oreq)
		if err != nil {
			fmt.Printf("Failed to grant access to repo %d to user %s: %s\n", crl.RepositoryID, userid, errors.ErrorString(err))
		} else {
			fmt.Printf("Objectauth granted access to repo #%d to user %s\n", crl.RepositoryID, userid)
		}
	}

	if req.CompleteForAccess {
		sr.RunPostReceive = true
		sr.RunPreReceive = true
		sr.CreatedComplete = true
		sr.Deleted = false
		sr.Forking = false
		sr.ReadOnly = false
		sr.DenyMessage = ""
		if sr.CreateUser == "" {
			sr.CreateUser = u.ID
		}
		err = db.DefaultDBSourceRepository().Update(ctx, sr)
		if err != nil {
			fmt.Printf("Failed to update: %s\n", errors.ErrorString(err))
			return nil, err
		}
	}
	return sr, nil
}

func (g *GIT2) SetRepoFlags(ctx context.Context, req *gitpb.SetRepoFlagsRequest) (*common.Void, error) {
	gr, err := g.repo_store.ByID(ctx, req.RepoID)
	if err != nil {
		return nil, err
	}

	err = isDeleted(ctx, gr)
	if err != nil {
		return nil, err
	}

	gr.RunPostReceive = req.RunPostReceive
	gr.RunPreReceive = req.RunPreReceive
	gr.ReadOnly = req.ReadOnly
	err = g.repo_store.Update(ctx, gr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Repository: %d (%s) - Postreceive Hook set to %v, PreReceive Hook set to %v\n", gr.ID, gr.ArtefactName, gr.RunPostReceive, gr.RunPreReceive)
	return &common.Void{}, nil
}

func req2repo(req *gitpb.CreateRepoRequest) *gitpb.SourceRepository {
	temp := utils.RandomString(60)
	sr := &gitpb.SourceRepository{
		FilePath:       temp,
		ArtefactName:   req.ArtefactName,
		RunPostReceive: true,
		RunPreReceive:  true,
		Description:    req.Description,
	}
	return sr
}

/*
we end up here if we are asked to create a repository under an existing URL
*/
func (g *GIT2) create_again(ctx context.Context, url *gitpb.SourceRepositoryURL, req *gitpb.CreateRepoRequest) (*gitpb.SourceRepository, error) {
	sr, err := g.repo_store.ByID(ctx, url.V2RepositoryID)
	if err != nil {
		return nil, err
	}
	// check if repo matches previous create attempt
	nr := req2repo(req)
	nome := errors.InvalidArgs(ctx, "url already exists", "url already used %s/%s", req.URL.Host, req.URL.Path)
	if nr.ArtefactName != sr.ArtefactName {
		return nil, nome
	}
	if nr.Description != sr.Description {
		return nil, nome
	}
	return sr, nil

}
func (g *GIT2) RepoBuilderComplete(ctx context.Context, req *gitpb.ByIDRequest) (*common.Void, error) {
	u := auth.GetService(ctx)
	if u == nil {
		return nil, errors.AccessDenied(ctx, "repobuilder only")
	}
	REPO_SERVICE_ID := auth.GetServiceIDByName("repobuilder.RepoBuilder")
	if u.ID != REPO_SERVICE_ID {
		fmt.Printf("UserID: %s\n", u.ID)
		return nil, errors.AccessDenied(ctx, "repobuilder service only")
	}
	gr, err := g.repo_store.ByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Setting Repository %d (%s) to 'repobuildercomplete'\n", gr.ID, gr.ArtefactName)
	gr.CreatedComplete = true
	err = g.repo_store.Update(ctx, gr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Repository: %d (%s) - RepoBuilder complete\n", req.ID, gr.ArtefactName)
	return &common.Void{}, nil

}

func (g *GIT2) ResetRepository(ctx context.Context, req *gitpb.ByIDRequest) (*common.Void, error) {
	fmt.Printf("Request to reset repository #%d\n", req.ID)
	if !isrepobuilder(ctx) {
		return nil, errors.AccessDenied(ctx, "requires repobuild account")
	}
	ok := false
	for _, r := range RESETTABLE_REPOS {
		if r == req.ID {
			ok = true
			break
		}
	}

	repo, err := g.repo_store.ByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		// in addition to the specifically listed ones, we also allow those ones to be reset:
		if !repo.CreatedComplete || repo.UserCommits == 0 {
			ok = true
		}
	}
	if !ok {
		return nil, errors.InvalidArgs(ctx, "invalid repo id", "invalid repo id %d", req.ID)
	}

	err = isDeleted(ctx, repo)
	if err != nil {
		return nil, err
	}

	urls, err := g.url_store.ByV2RepositoryID(ctx, repo.ID)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, errors.NotFound(ctx, "no source repositorylocation")
	}
	surl := urls[0]

	repo.CreatedComplete = false
	err = g.repo_store.Update(ctx, repo)
	if err != nil {
		return nil, err
	}
	ctxs, err := auth.SerialiseContextToString(ctx)
	if err != nil {
		return nil, err
	}
	crl := &gitpb.CreateRepoLog{
		RepositoryID:     repo.ID,
		UserID:           "",
		Context:          ctxs,
		Action:           1,
		Started:          uint32(time.Now().Unix()),
		AssociationToken: utils.RandomString(256),
	}
	_, err = g.repocreatelog_store.Save(ctx, crl)
	if err != nil {
		return nil, err
	}

	h := http.HTTP{}
	url := fmt.Sprintf("https://%s/git/self/Recreate", surl.Host)
	r := utils.RandomString(48)
	h.SetHeader(REPEAT_BACK_HEADER, r)
	h.SetHeader("X-AssociationToken", crl.AssociationToken)
	hb := h.Get(url)
	err = hb.Error()
	if err != nil {
		fmt.Printf("HTTP-Recreate request failed: %s\n", utils.ErrorString(err))
		return nil, err
	}
	body := string(hb.Body())
	if !strings.Contains(body, r) {
		fmt.Printf("Did not reset repo at %s: (%s)\n", url, body)
		return nil, errors.Errorf("failed to recreate: not a gitserver reply")
	}
	fmt.Printf("Resetted repo at %s\n", url)
	return &common.Void{}, nil
}
func (g *GIT2) DeleteRepository(ctx context.Context, req *gitpb.ByIDRequest) (*common.Void, error) {
	sr, err := g.RepoByID(ctx, req)
	if err != nil {
		return nil, err
	}
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "need user", "need user")
	}
	sr.Deleted = true
	sr.DeletedTimestamp = uint32(time.Now().Unix())
	sr.DeleteUser = u.ID
	err = g.repo_store.Update(ctx, sr)
	if err != nil {
		return nil, err
	}
	return &common.Void{}, nil
}
func (g *GIT2) GetReposTags(ctx context.Context, req *gitpb.RepoTagRequest) (*gitpb.SourceRepositoryList, error) {
	return g.GetReposWithFilter(ctx, &gitpb.RepoFilter{Tags: req})
	/*
		sr, err := g.GetRepos(ctx, &common.Void{})
		if err != nil {
			return nil, err
		}
		var repos []*gitpb.SourceRepository
		// filter by tag..
		mask := uint32((1 << req.Tag))
		for _, r := range sr.Repos {
			//		fmt.Printf("Repo %03d: Tags: %d\n", r.ID, r.Tags)
			if (r.Tags & mask) > 0 {
				repos = append(repos, r)
			}
		}
		fmt.Printf("Of %d repos, %d had tag %d (%d)\n", len(sr.Repos), len(repos), req.Tag, mask)
		sr.Repos = repos
		return sr, nil
	*/
}
func isDeleted(ctx context.Context, sr *gitpb.SourceRepository) error {
	if !sr.Deleted {
		return nil
	}
	return errors.NotFound(ctx, "repo %d (%s) was deleted", sr.ID, sr.ArtefactName)
}
func (g *GIT2) CheckGitServer(ctx context.Context, req *gitpb.CheckGitRequest) (*gitpb.CheckGitResponse, error) {
	_, b, err := query.SendPing(ctx, req.Host)
	if err != nil {
		fmt.Printf("Check failed: %s\n", utils.ErrorString(err))
		b = false
	}
	cr := &gitpb.CheckGitResponse{Success: b}
	return cr, nil

}
func (g *GIT2) GetRecentBuilds(ctx context.Context, req *gitpb.ByIDRequest) (*gitpb.BuildList, error) {
	ot := &oa.AuthRequest{ObjectType: oa.OBJECTTYPE_GitRepository, ObjectID: req.ID}
	ol, err := oa.GetObjectAuthClient().AskObjectAccess(ctx, ot)
	if err != nil {
		return nil, err
	}
	if ol == nil {
		fmt.Printf("for a weird reason we did not get a permissions object but no error either")
		return nil, errors.Errorf("permission error")
	}
	if ol.Permissions == nil {
		return nil, errors.AccessDenied(ctx, "access denied to git repository %d (no permission)", req.ID)
	}
	if !auth.IsRoot(ctx) {
		if !ol.Permissions.View && !ol.Permissions.Read && !ol.Permissions.Execute {
			return nil, errors.AccessDenied(ctx, "access denied to git repository %d", req.ID)
		}
	}
	res := &gitpb.BuildList{}
	builds, err := g.build_store.ByRepositoryID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	sort.Slice(builds, func(i, j int) bool {
		return builds[i].Timestamp >= builds[j].Timestamp
	})
	res.Builds = builds
	return res, nil
}

// return error if host does not point to a git server
func checkValidHost(ctx context.Context, host string) error {
	_, b, _ := query.SendPing(ctx, host)
	if !b {
		return errors.NotFound(ctx, "no gitserver at host \"%s\"", host)
	}
	return nil
}

/*
 see readme.txt - this is called by the hook on the server
 gitclient->http->git-cgi.go->git-http-backend->hook->grpc.Connect("localhost").RunLocalHook(...)
*/

func (g *GIT2) RunLocalHook(req *gitpb.HookRequest, srv gitpb.GIT2_RunLocalHookServer) error {
	ld := crossprocdata.GetLocalData(req.RequestKey)
	if ld == nil {
		fmt.Printf("failed. no crossdata\n")
		return errors.Errorf("attempt to invoke hook failed, invalid or missing requestkey from gitprocess")
	}
	hr := ld.HTTPRequest.(*HTTPRequest)
	ctx := srv.Context()
	ch := checks.NewChecker(ctx, hr.repo.gitrepo, req.OldRev, req.NewRev)
	fmt.Printf("Serverprocess-space hook \"%s\" executing for repo #%d \"%s\"\n", req.HookName, hr.repo.gitrepo.ID, hr.repo.AbsDirectory())
	var err error
	if req.HookName == "update" {
		err = ch.OnUpdate(srv)
	} else if req.HookName == "post-update" {
		print_cross_proc_data(req, ld)
	} else {
		print_cross_proc_data(req, ld)
		fmt.Printf("Serverprocess-space hook \"%s\" ignored for repo #%d \"%s\"\n", req.HookName, hr.repo.gitrepo.ID, hr.repo.AbsDirectory())
		return nil
	}
	if err != nil {
		fmt.Printf("Hook \"%s\" failed: %s\n", req.HookName, err)
		srv.Send(&gitpb.HookResponse{ErrorMessage: fmt.Sprintf("%s", err)})
		return nil
	}

	fmt.Printf("Serverprocess-space hook \"%s\" completed for repo #%d \"%s\"\n", req.HookName, hr.repo.gitrepo.ID, hr.repo.AbsDirectory())
	return nil
}
func (g *GIT2) GetLatestSuccessfulBuild(ctx context.Context, req *gitpb.ByIDRequest) (*gitpb.Build, error) {
	err := wants_access_to_build(ctx, req.ID) // despite the name, input *is* a repositoryid
	if err != nil {
		return nil, err
	}
	//	builds, err := g.build_store.FromQuery(ctx, " repositoryid=$1 and success = true order by id desc limit 1", req.ID)
	q := g.build_store.NewQuery()
	q.AddEqual("repositoryid", req.ID)
	q.AddEqual("success", true)
	q.Limit(1)
	q.OrderBy("id desc")
	builds, err := g.build_store.ByDBQuery(ctx, q)

	if err != nil {
		return nil, err
	}
	if len(builds) == 0 {
		return nil, errors.NotFound(ctx, "no build for repo %d", req.ID)
	}
	res := builds[0]
	return res, nil
}
func (g *GIT2) GetLatestBuild(ctx context.Context, req *gitpb.ByIDRequest) (*gitpb.Build, error) {
	err := wants_access_to_build(ctx, req.ID) // despite the name, input *is* a repositoryid
	if err != nil {
		return nil, err
	}

	q := g.build_store.NewQuery()
	q.AddEqual("repositoryid", req.ID)
	q.Limit(1)
	q.OrderBy("id desc")
	builds, err := g.build_store.ByDBQuery(ctx, q)
	//	ctx, " repositoryid=$1 order by id desc limit 1", req.ID)
	if err != nil {
		return nil, err
	}
	if len(builds) == 0 {
		return nil, errors.NotFound(ctx, "no build for repo %d", req.ID)
	}
	res := builds[0]
	return res, nil
}
func (g *GIT2) FindRepoByURL(ctx context.Context, req *gitpb.ByURLRequest) (*gitpb.SourceRepositoryResponse, error) {
	//fmt.Printf("Finding repo by url \"%s\"\n", req.URL)
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, err
	}
	path := u.Path
	path = strings.TrimPrefix(path, "/git/")
	urls, err := g.url_store.ByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	host := u.Host
	id := uint64(0)
	for _, u := range urls {
		if u.Host == host {
			id = u.V2RepositoryID
		}
	}
	res := &gitpb.SourceRepositoryResponse{Found: false}
	if id == 0 {
		//		fmt.Printf("no repo for Host: \"%s\", Path: \"%s\"\n", host, path)
		return res, nil
	}
	r := &gitpb.ByIDRequest{ID: id}
	repo, err := g.RepoByID(ctx, r)
	if err != nil {
		return nil, err
	}
	res.Repository = repo
	res.Found = true
	//	fmt.Printf("Repo \"%s\" - #%d\n", req.URL, res.Repository.ID)
	return res, nil
}
func (g *GIT2) GetNumberCommitsUser(ctx context.Context, req *gitpb.NumberCommitsUserRequest) (*gitpb.NumberCommitsUserResponse, error) {
	commits, err := db.DefaultDBGitAccessLog().ByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	val := uint32(0)
	for _, c := range commits {
		if c.Timestamp > req.Timestamp {
			continue
		}
		if !c.Write {
			continue
		}
		val++
	}
	return &gitpb.NumberCommitsUserResponse{Commits: val}, nil
}
func print_cross_proc_data(req *gitpb.HookRequest, ld *crossprocdata.LocalData) {
	if !*debug_githook {
		return
	}
	fmt.Printf("Args:\n")
	for _, a := range req.Args {
		fmt.Printf("   Arg: \"%s\"\n", a)
	}
	fmt.Printf("Environment:\n")
	for _, a := range req.Environment {
		fmt.Printf("   Env: \"%s\"\n", a)
	}
}
