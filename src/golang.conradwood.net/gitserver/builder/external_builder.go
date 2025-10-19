package builder

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"

	"golang.conradwood.net/apis/gitbuilder"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/artefacts"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
)

var (
	def_routing = flag.Bool("use_default_routing_tags", true, "if true use default routing tags if none is specified for a repository")
)

// implementation normally GitTrigger
type ExternalGitTrigger interface {
	RepositoryID() uint64
	//ArtefactID() uint64
	NewRev() string
	Branch() string
	UserID() string
	ExcludeBuildScripts() []string
}

func getartefactid(ctx context.Context, repo *gitpb.SourceRepository) (uint64, error) {
	afc, err := artefacts.CreateIfRequired(ctx, repo)
	if err != nil {
		return 0, err
	}
	return afc.Meta.ID, nil
}

func RunExternalBuilder(ctx context.Context, gt ExternalGitTrigger, buildid uint64, w io.Writer) (*gitbuilder.BuildResponse, error) {
	repo, err := db.DefaultDBSourceRepository().ByID(ctx, gt.RepositoryID())
	if err != nil {
		return nil, err
	}

	urls, err := db.DefaultDBSourceRepositoryURL().ByV2RepositoryID(ctx, gt.RepositoryID())
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("Repository %d (%s) has no urls\n", repo.ID, repo.ArtefactName)
	}
	repo.URLs = urls
	url := fmt.Sprintf("https://%s/git/%s", urls[0].Host, urls[0].Path)
	afcid, err := getartefactid(ctx, repo)
	if err != nil {
		return nil, err
	}
	gb := gitbuilder.GetGitBuilderClient()
	br := &gitbuilder.BuildRequest{
		GitURL:              url,
		CommitID:            gt.NewRev(),
		BuildNumber:         buildid,
		RepositoryID:        gt.RepositoryID(),
		RepoName:            repo.ArtefactName,
		ArtefactName:        repo.ArtefactName,
		ExcludeBuildScripts: gt.ExcludeBuildScripts(),
		ArtefactID:          afcid,
		//ArtefactID:          gt.ArtefactID(),
	}
	// might have to add special routing tags to context to route it to a SPECIFIC builder
	rm := make(map[string]string)
	if repo.BuildRoutingTagName != "" && repo.BuildRoutingTagValue != "" {
		rm[repo.BuildRoutingTagName] = repo.BuildRoutingTagValue
	} else {
		if *def_routing {
			rm["provides"] = "default"
		}
	}
	ctx = authremote.DerivedContextWithRouting(ctx, rm, true)
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "no user for builder")
	}
	cl, err := gb.Build(ctx, br)
	if err != nil {
		return nil, err
	}
	var lastResponse *gitbuilder.BuildResponse
	for {
		res, err := cl.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if res.Complete {
			lastResponse = res
		}
		if len(res.Stdout) != 0 {
			_, err = w.Write(res.Stdout)
			if err != nil {
				return lastResponse, err
			}
		}
	}

	if lastResponse == nil || !lastResponse.Success {
		return lastResponse, fmt.Errorf("build failed (%s)", lastResponse.ResultMessage)
	}
	return lastResponse, nil

}

// use the gitbuilder service to build instead of locally forking it off
// e.g. ref== 'master', newrev == commitid
func external_builder(ctx context.Context, gt ExternalGitTrigger, w io.Writer) error {
	//	gi := gt.gitinfo

	psql, err := sql.Open()
	if err != nil {
		return err
	}
	// create build
	bdb := db.NewDBBuild(psql)
	nb := &gitpb.Build{
		UserID:       gt.UserID(),
		RepositoryID: gt.RepositoryID(),
		CommitHash:   gt.NewRev(),
		Branch:       gt.Branch(),
		LogMessage:   "logmessage unavailable", // to get the logmessage we have to check the repo out
		Timestamp:    uint32(time.Now().Unix()),
	}

	id, err := bdb.Save(context.Background(), nb)
	if err != nil {
		fmt.Printf("Failed to save to database: %s\n", err)
		return err
	}

	repodb := db.DefaultDBSourceRepository()
	repo, err := repodb.ByID(ctx, gt.RepositoryID())
	if err != nil {
		return err
	}

	//run external builder..
	urls, err := db.DefaultDBSourceRepositoryURL().ByV2RepositoryID(ctx, gt.RepositoryID())
	if err != nil {
		return err
	}
	if len(urls) == 0 {
		return fmt.Errorf("Repository %d (%s) has no urls\n", repo.ID, repo.ArtefactName)
	}
	repo.URLs = urls
	url := fmt.Sprintf("https://%s/git/%s", urls[0].Host, urls[0].Path)
	afcid, err := getartefactid(ctx, repo)
	if err != nil {
		return err
	}

	gb := gitbuilder.GetGitBuilderClient()
	br := &gitbuilder.BuildRequest{
		GitURL:       url,
		CommitID:     gt.NewRev(),
		BuildNumber:  id,
		RepositoryID: gt.RepositoryID(),
		RepoName:     repo.ArtefactName,
		ArtefactName: repo.ArtefactName,
		ArtefactID:   afcid,
		//		ArtefactID:   gt.ArtefactID(),
	}
	// might have to add special routing tags to context to route it to a SPECIFIC builder
	rm := make(map[string]string)
	if repo.BuildRoutingTagName != "" && repo.BuildRoutingTagValue != "" {
		rm[repo.BuildRoutingTagName] = repo.BuildRoutingTagValue
	} else {
		if *def_routing {
			rm["provides"] = "default"
		}
	}
	ctx = authremote.DerivedContextWithRouting(ctx, rm, true)
	cl, err := gb.Build(ctx, br)
	if err != nil {
		return err
	}
	var lastResponse *gitbuilder.BuildResponse
	for {
		res, err := cl.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if res.Complete {
			lastResponse = res
		}
		if len(res.Stdout) != 0 {
			_, err = w.Write(res.Stdout)
			if err != nil {
				return err
			}
		}
	}

	if lastResponse == nil {
		fmt.Printf("No final response from gitbuilder received. build information might be incomplete\n")
	} else if lastResponse != nil {
		nb.Success = lastResponse.Success
		nb.LogMessage = lastResponse.LogMessage
		fmt.Printf("Gitbuilder success: %v\n", nb.Success)
		fmt.Printf("Gitbuilder logmessage:\n%s\n", nb.LogMessage)
		err := bdb.Update(ctx, nb)
		if err != nil {
			fmt.Printf("Failed to set logmessage: %s\n", err)
		}
	}

	if lastResponse == nil || !lastResponse.Success {
		return fmt.Errorf("build failed (%s)", lastResponse.ResultMessage)
	}
	return nil
}
