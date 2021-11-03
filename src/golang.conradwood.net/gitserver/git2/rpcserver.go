package git2

import (
	"context"
	"fmt"
	"golang.conradwood.net/apis/common"
	gitpb "golang.conradwood.net/apis/gitserver"
	oa "golang.conradwood.net/apis/objectauth"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/gitserver/query"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/cache"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/http"
	"golang.conradwood.net/go-easyops/utils"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	REPO_SERVICE_ID = "3539"
	WEB_SERVICE_ID  = "145"
)

var (
	// these repos may be 'reset' back to 'bare' at any time
	RESETTABLE_REPOS = []uint64{216}
	urlCache         = cache.NewResolvingCache("giturlcache", time.Duration(15)*time.Minute, 1000)
)

type GIT2 struct {
	url_store           *db.DBSourceRepositoryURL
	repo_store          *db.DBSourceRepository
	repocreatelog_store *db.DBCreateRepoLog
	build_store         *db.DBBuild
}

func (g *GIT2) init() error {
	g.url_store = db.NewDBSourceRepositoryURL(psql)
	g.repo_store = db.NewDBSourceRepository(psql)
	g.repocreatelog_store = db.NewDBCreateRepoLog(psql)
	g.build_store = db.NewDBBuild(psql)
	return nil

}
func (g *GIT2) RepoByURL(ctx context.Context, req *gitpb.ByURLRequest) (*gitpb.SourceRepository, error) {
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
	if id == 0 {
		fmt.Printf("Host: \"%s\", Path: \"%s\"\n", host, path)
		return nil, errors.NotFound(ctx, "no such repo")
	}
	r := &gitpb.ByIDRequest{ID: id}
	return g.RepoByID(ctx, r)
}
func (g *GIT2) RepoByID(ctx context.Context, req *gitpb.ByIDRequest) (*gitpb.SourceRepository, error) {
	fmt.Printf("Getting repo by id %d\n", req.ID)
	ot := &oa.AuthRequest{ObjectType: oa.OBJECTTYPE_GitRepository, ObjectID: req.ID}
	ol, err := oa.GetObjectAuthClient().AskObjectAccess(ctx, ot)
	if err != nil {
		return nil, err
	}
	if ol == nil {
		fmt.Printf("for a weird reason we did not get a permissions object but no error either")
		return nil, fmt.Errorf("permission error")
	}
	if ol.Permissions == nil {
		return nil, errors.AccessDenied(ctx, "access denied to git repository %d (no permission)", req.ID)
	}
	if !auth.IsRoot(ctx) {
		if !ol.Permissions.View && !ol.Permissions.Read && !ol.Permissions.Execute {
			return nil, errors.AccessDenied(ctx, "access denied to git repository %d", req.ID)
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
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "no user: no repos")
	}
	fmt.Printf("GetRepos() - got %d repos\n", len(repos))
	res := &gitpb.SourceRepositoryList{}
	ot := &oa.ObjectType{ObjectType: oa.OBJECTTYPE_GitRepository}
	ol, err := oa.GetObjectAuthClient().AvailableObjects(ctx, ot)
	if err != nil {
		return nil, err
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
		for _, id := range ol.ObjectIDs {
			if id == sr.ID {
				include = true
				break
			}
		}
		if include || *disable_access_check {
			res.Repos = append(res.Repos, sr)

			uao, err := urlCache.Retrieve(fmt.Sprintf("%d", sr.ID), func(k string) (interface{}, error) {
				return g.url_store.ByV2RepositoryID(ctx, sr.ID)
			})
			if err != nil {
				return nil, err
			}
			sr.URLs = uao.([]*gitpb.SourceRepositoryURL)
		}
	}
	fmt.Printf("GetRepos() - returning %d repos to user %s (%s)\n", len(res.Repos), auth.Description(u), u.ID)
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
	h.SetHeader("X-AssociationToken", crl.AssociationToken)
	hb := h.Get(url)
	err = hb.Error()
	if err != nil {
		fmt.Printf("HTTP-Recreate request failed: %s\n", utils.ErrorString(err))
		return nil, err
	}
	fmt.Printf("Resetted repo at %s: (%s)\n", url, string(hb.Body()))
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
		return nil, fmt.Errorf("permission error")
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
		return errors.NotFound(ctx, "no gitserver at that host", "no gitserver at host \"%s\"", host)
	}
	return nil
}

func (g *GIT2) GetLatestBuild(ctx context.Context, req *gitpb.ByIDRequest) (*gitpb.Build, error) {
	svc := auth.GetService(ctx)
	check_user := true
	if svc != nil && svc.ID == WEB_SERVICE_ID {
		check_user = false
	}

	if check_user {
		ot := &oa.AuthRequest{ObjectType: oa.OBJECTTYPE_GitRepository, ObjectID: req.ID}
		ol, err := oa.GetObjectAuthClient().AskObjectAccess(ctx, ot)
		if err != nil {
			return nil, err
		}
		if ol == nil {
			fmt.Printf("for a weird reason we did not get a permissions object but no error either")
			return nil, fmt.Errorf("permission error")
		}
		if ol.Permissions == nil {
			return nil, errors.AccessDenied(ctx, "access denied to git repository %d (no permission)", req.ID)
		}
		if !ol.Permissions.Read {
			return nil, errors.AccessDenied(ctx, "access denied to git repository %d (no read permission)", req.ID)
		}
	}
	builds, err := g.build_store.FromQuery(ctx, " repositoryid=$1 order by id desc limit 1", req.ID)
	if err != nil {
		return nil, err
	}
	if len(builds) == 0 {
		return nil, errors.NotFound(ctx, "no build for repo %d", req.ID)
	}
	return builds[0], nil
}
