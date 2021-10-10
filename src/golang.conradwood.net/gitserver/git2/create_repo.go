package git2

import (
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/tokens"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"time"
)

func (h *HTTPRequest) RecreateRepo() {
	crp, err := h.GetCreateLog()
	if err != nil {
		h.Error(err)
		return
	}

	ctx := tokens.ContextWithToken()
	if crp.Finished != 0 || crp.Success {
		fmt.Printf("Repolog %d reused. action not taken\n", crp.ID)
		h.Error(errors.InvalidArgs(ctx, "association token not or no longer valid", "association token refered to log %d which is completed already", crp.ID))
		return
	}

	repo, err := h.git2.repo_store.ByID(ctx, crp.RepositoryID)
	if err != nil {
		fmt.Printf("failed to load repository: %s\n", err)
		h.Error(err)
		return
	}
	h.repo = &Repo{gitrepo: repo}
	td := h.repo.AbsDirectory()
	os.RemoveAll(td)
	_, err = h.CreateRepoWithError(false)
	if err != nil {
		h.Error(err)
		return
	}
	crp.Finished = uint32(time.Now().Unix())
	crp.Success = true
	err = h.git2.repocreatelog_store.Update(ctx, crp)
	if err != nil {
		fmt.Printf("Failed to update database: %s\n", err)
		h.Error(err)
		return
	}

}
func (h *HTTPRequest) CreateRepo() {
	crp, err := h.CreateRepoWithError(true)
	ctx := tokens.ContextWithToken()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		h.Error(err)
		if crp != nil {
			crp.Success = false
			crp.ErrorMessage = utils.ErrorString(err)
			crp.Finished = uint32(time.Now().Unix())
			err = h.git2.repocreatelog_store.Update(ctx, crp)
			if err != nil {
				fmt.Printf("Failed to update database: %s\n", err)
				h.Error(err)
				return
			}
		}
		return
	}

	if crp.Finished == 0 || !crp.Success {
		crp.Finished = uint32(time.Now().Unix())
		crp.Success = true
		err = h.git2.repocreatelog_store.Update(ctx, crp)
		if err != nil {
			fmt.Printf("Failed to update database: %s\n", err)
			h.Error(err)
			return
		}
	}

	h.w.Write([]byte("Repository created\n"))
	fmt.Printf("Repository created.\n")
}
func (h *HTTPRequest) CreateRepoWithError(requireUser bool) (*gitpb.CreateRepoLog, error) {
	fmt.Printf("Creating repository %s\n", h.gurl.RepoPath())
	crp, err := h.GetCreateLog()
	if err != nil {
		return nil, err
	}
	if crp.Success {
		fmt.Printf("Already processed CreateRepoLog #%d successfully\n", crp.ID)
		return crp, nil
	}
	fmt.Printf("Creating Repository %d\n", crp.RepositoryID)
	ctx, err := auth.RecreateContextWithTimeout(time.Duration(5)*time.Minute, []byte(crp.Context))
	if err != nil {
		fmt.Printf("Context is not valid: %s\n", err)
		return crp, err
	}
	h.user = auth.GetUser(ctx)

	if h.user == nil && requireUser {
		fmt.Printf("No user in deserialised context!")
		return crp, fmt.Errorf("invalid unauthenticated request")
	}

	repo, err := h.git2.repo_store.ByID(ctx, crp.RepositoryID)
	if err != nil {
		fmt.Printf("failed to load repository: %s\n", err)
		return crp, err
	}
	h.repo = &Repo{gitrepo: repo}
	td := h.repo.AbsDirectory()
	if utils.FileExists(td + "/config") {
		fmt.Printf("There is already a repository at %s\n", td)
		return crp, fmt.Errorf("repository exists already")
	}
	fmt.Printf("Need to create repo at: %s\n", td)
	os.MkdirAll(td, 0777)
	if repo.Forking {
		src, err := h.git2.repo_store.ByID(ctx, repo.ForkedFrom)
		if err != nil {
			fmt.Printf("failed to load source repository: %s\n", err)
			return crp, err
		}
		copyrepo(src, h.repo.gitrepo)
	} else {
		l := linux.New()
		out, err := l.SafelyExecuteWithDir([]string{"git", "--bare", "init"}, td, nil)
		if err != nil {
			fmt.Printf("git bare init failed: %s\n", out)
			return nil, fmt.Errorf("failed to init git repository")
		}
	}
	return crp, nil
}

// load the createlog from url/headers
func (h *HTTPRequest) GetCreateLog() (*gitpb.CreateRepoLog, error) {
	actok := h.SingleHeader("x-associationtoken")
	if actok == "" {
		return nil, fmt.Errorf("invalid unauthenticated request")
	}

	// find access token in db
	ctx := tokens.ContextWithToken()
	crpx, err := h.git2.repocreatelog_store.ByAssociationToken(ctx, actok)
	if err != nil {
		fmt.Printf("Failed to get associationtoken \"%s\" from db: %s\n", actok, err)
		return nil, fmt.Errorf("invalid unauthenticated request")
	}
	if len(crpx) == 0 {
		fmt.Printf("No database entry for associationtoken \"%s\"\n", actok)
		return nil, fmt.Errorf("invalid authenticationtoken")
	}
	crp := crpx[0]
	return crp, nil
}
