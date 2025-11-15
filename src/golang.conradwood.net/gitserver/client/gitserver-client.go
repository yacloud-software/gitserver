package main

import (
	"flag"
	"fmt"
	"sort"

	"golang.conradwood.net/apis/common"
	pb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/artefacts"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"

	//	"golang.conradwood.net/go-easyops/client"
	"io"
	"os"
	"strings"
	"time"

	"golang.conradwood.net/go-easyops/ctx"
	"golang.conradwood.net/go-easyops/utils"
)

const (
	FORMAT_HUMAN = 1
	FORMAT_SHELL = 2
)

var (
	format      = flag.String("of", "human", "output format. [human|bash]")
	githost     = flag.String("githost", "git.conradwood.net", "default git host for create")
	exc         = flag.String("exc_scripts", "", "comma delimeted list of scripts to be excluded from run. passed on to gitbuilder. typically 'DIST'")
	latest      = flag.Bool("latest", false, "get latest build of current repo")
	desc        = flag.String("desc", "", "description of new repo")
	delete      = flag.Bool("delete", false, "delete repo")
	print_repos = flag.Bool("print_repos", false, "print repositories")
	fork        = flag.Bool("fork", false, "fork a repo")
	create      = flag.Bool("create", false, "create a repo")
	aname       = flag.String("artefactname", "", "artefactname")
	repoid      = flag.Int("repoid", 0, "repo id to operate on")
	repourl     = flag.String("repourl", "", "repo url to operate on")
	print_desc  = flag.Bool("print_desc", false, "if true print repository description")
	print_cpu   = flag.Bool("print_cpu", true, "print cpu utilisation")
	host        = flag.String("host", "", "hostname to serve repo on")
	path        = flag.String("path", "", "pathname to serve repo on (the repository name as far as git is concerned)")
	check       = flag.Bool("check", false, "if true, check if there is a git server at host")
	getrepo     = flag.Bool("info", false, "if true get repo information (requires repoid)")
	debug       = flag.Bool("debug", false, "debug mode")
	builds      = flag.Bool("builds", false, "do builds")
	rebuild     = flag.Uint64("rebuild", 0, "trigger a build")
	denymsg     = flag.Bool("denymsg", false, "if set, disable a repo with the message. ")
	message     = flag.String("message", "", "the message to set or clear")
	build       = flag.Uint64("build", 0, "build to operate on")
	set_tag     = flag.Uint64("tag_set", 0, "set tag on a given build")
	get_tag     = flag.Uint64("tag_get", 0, "get tag from a repo")
)

func main() {
	flag.Parse()
	if *set_tag != 0 {
		utils.Bail("failed to set deny message: ", SetTag())
		os.Exit(0)
	}
	if *get_tag != 0 {
		utils.Bail("failed to set deny message: ", GetTag())
		os.Exit(0)
	}
	if *denymsg {
		utils.Bail("failed to set deny message: ", denyMsg())
		os.Exit(0)
	}
	if *print_desc {
		utils.Bail("failed to print description: ", printDesc())
		os.Exit(0)
	}
	if *rebuild != 0 {
		utils.Bail("rebuild failed", Rebuild())
		os.Exit(0)
	}
	if *builds {
		DoBuilds()
		os.Exit(0)
	}
	if *latest {
		Latest()
		os.Exit(0)
	}
	ctx := authremote.Context()
	if *check {
		Check()
		os.Exit(0)
	}
	if *delete {
		Delete()
		os.Exit(0)
	}
	if *fork {
		Fork()
		os.Exit(0)
	}
	if *create {
		Create()
		os.Exit(0)
	}
	if *print_repos {
		rl, err := pb.GetGIT2Client().GetRepos(ctx, &common.Void{})
		utils.Bail("failed to get repos", err)
		urlCounter := 0
		for _, r := range rl.Repos {
			fmt.Printf("Repository: #%02d %s\n", r.ID, r.ArtefactName)
			for _, u := range r.URLs {
				urlCounter++
				fmt.Printf("       URL: https://%s/%s\n", u.Host, u.Path)
			}
		}
		fmt.Printf("%d repos, %d urls\n", len(rl.Repos), urlCounter)
	}
	if *getrepo {
		utils.Bail("-info failed", showrepo())
		os.Exit(0)
	}
	fmt.Printf("Done.\n")
	os.Exit(0)
}
func showrepo() error {
	ctx := authremote.Context()
	rl := GetRepository(ctx)
	af, aerr := artefacts.RepositoryIDToArtefactID(rl)
	afs := "n/a"
	if aerr != nil {
		fmt.Printf("failed to get artefact id: %s\n", utils.ErrorString(aerr))
	} else {
		afs = fmt.Sprintf("%d", af)
	}
	rid := &pb.ByIDRequest{ID: rl.ID}
	if Format() == FORMAT_HUMAN {
		fmt.Printf("RepositoryID   : %d\n", rl.ID)
		fmt.Printf("ArtefactID     : %s\n", afs)
		fmt.Printf("Artefact       : %s\n", rl.ArtefactName)
		fmt.Printf("Last Commit    : %s\n", utils.TimestampString(rl.LastCommit))
		fmt.Printf("Last Committer : %s\n", rl.LastCommitUser)
		for _, u := range rl.URLs {
			url := fmt.Sprintf("https://%s/git/%s", u.Host, u.Path)
			fmt.Printf("URL            : %s\n", url)
		}
		b, err := pb.GetGIT2Client().GetLatestBuild(ctx, rid)
		utils.Bail("did not get latest build", err)
		bs, err := pb.GetGIT2Client().GetLatestSuccessfulBuild(ctx, rid)
		utils.Bail("did not get latest build", err)

		s := fmt.Sprintf("%d", b.ID)
		if bs.ID != b.ID {
			s = fmt.Sprintf("%d (last successful build: %d)", b.ID, bs.ID)
		}
		fmt.Printf("Latest Build   : %s\n", s)
	} else if Format() == FORMAT_SHELL {
		b, err := pb.GetGIT2Client().GetLatestBuild(ctx, rid)
		utils.Bail("did not get latest build", err)
		bs, err := pb.GetGIT2Client().GetLatestSuccessfulBuild(ctx, rid)
		utils.Bail("did not get latest build", err)
		fmt.Printf("GITSERVER_REPOSITORYID=%d\n", rl.ID)
		fmt.Printf("GITSERVER_ARTEFACTID=%s\n", afs)
		fmt.Printf("GITSERVER_ARTEFACT=%s\n", rl.ArtefactName)
		fmt.Printf("GITSERVER_LAST_COMMIT=%d\n", rl.LastCommit)
		fmt.Printf("GITSERVER_LAST_COMMIT_STRING=\"%s\"\n", utils.TimestampString(rl.LastCommit))
		fmt.Printf("GITSERVER_LAST_COMMITTER=%s\n", rl.LastCommitUser)
		sort.Slice(rl.URLs, func(i, j int) bool {
			s1 := rl.URLs[i].Host + rl.URLs[i].Path
			s2 := rl.URLs[j].Host + rl.URLs[j].Path
			return s1 < s2
		})
		u := ""
		url_host := ""
		if len(rl.URLs) > 0 {
			url_host = rl.URLs[0].Host
			u = rl.URLs[0].Host + "/" + rl.URLs[0].Path
		}
		fmt.Printf("GITSERVER_URL=\"https://%s\"\n", u)
		fmt.Printf("GITSERVER_URL_HOST=\"%s\"\n", url_host)
		fmt.Printf("GITSERVER_LATEST_BUILD=%d\n", b.ID)
		fmt.Printf("GITSERVER_LATEST_SUCCESSFUL_BUILD=%d\n", bs.ID)
	}

	return aerr
}

func Delete() {
	ctx := authremote.Context()
	fr := &pb.ByIDRequest{ID: uint64(*repoid)}
	_, err := pb.GetGIT2Client().DeleteRepository(ctx, fr)
	utils.Bail("Failed to delete()", err)
	fmt.Printf("deleted repository #%d\n", fr.ID)
}
func Check() {
	ctx := authremote.Context()
	fr := &pb.CheckGitRequest{Host: *host}
	cr, err := pb.GetGIT2Client().CheckGitServer(ctx, fr)
	utils.Bail("Failed to check()", err)
	fmt.Printf("checked repository %#v\n", cr)
}

func Fork() {
	ctx := authremote.Context()
	fr := &pb.ForkRequest{
		RepositoryID: uint64(*repoid),
		ArtefactName: *aname,
		URL:          &pb.SourceRepositoryURL{Host: *host, Path: *path},
	}
	rl, err := pb.GetGIT2Client().Fork(ctx, fr)
	utils.Bail("Failed to fork()", err)
	if Format() == FORMAT_HUMAN {
		fmt.Printf("Forked into #%d\n", rl.ID)
		fmt.Printf("Served at https://%s/git/%s\n", rl.URLs[0].Host, rl.URLs[0].Path)
	} else if Format() == FORMAT_SHELL {
		fmt.Printf("GITSERVER_NEW_REPOSITORYID=%d\n", rl.ID)
		fmt.Printf("GITSERVER_NEW_URL=\"https://%s/git/%s\"\n", rl.URLs[0].Host, rl.URLs[0].Path)
	} else {
		panic("inv format")
	}
}
func Create() {
	ctx := authremote.Context()
	u := auth.GetUser(ctx)
	if u == nil {
		fmt.Printf("Could not determine your useraccount.\n")
		os.Exit(10)
	}
	path := fmt.Sprintf("%s/%s.git", u.Abbrev, *aname)
	url := &pb.SourceRepositoryURL{Host: *githost, Path: path}
	fr := &pb.CreateRepoRequest{CompleteForAccess: true, ArtefactName: *aname, URL: url, Description: *desc}
	rl, err := pb.GetGIT2Client().CreateRepo(ctx, fr)
	utils.Bail("Failed to create repo", err)
	fmt.Printf("Created repository: %d at https://%s/%s\n", rl.ID, url.Host, url.Path)
}

func Rebuild() error {
	//	client.GetSignatureFromAuth()
	u, _ := authremote.GetLocalUsers()
	if u == nil {
		fmt.Printf("No local user.\n")
		os.Exit(10)
	}
	fmt.Printf("Running rebuild as you\n")
	auth.PrintSignedUser(u)
	cb := ctx.NewContextBuilder()
	cb.WithUser(u)
	cb.WithTimeout(time.Duration(60) * time.Minute)
	ctx := cb.ContextWithAutoCancel()
	br := &pb.RebuildRequest{
		ID:                  *rebuild,
		ExcludeBuildScripts: parseCommaDelimetedList(*exc),
	}
	rl, err := pb.GetGIT2Client().Rebuild(ctx, br)
	if err != nil {
		return err
	}
	failed := false
	var ferr error
	for {
		hr, err := rl.Recv()
		if hr != nil {
			if hr.Output != "" {
				fmt.Print(hr.Output)
			}
			if hr.ErrorMessage != "" {
				fmt.Println("************** ERROR")
				fmt.Println(hr.ErrorMessage)
				failed = true
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			failed = true
			ferr = err
			break
		}
	}
	if ferr != nil {
		return ferr
	}
	if failed {
		return fmt.Errorf("failed (errormessage in log)")
	}
	fmt.Printf("Build successful\n")
	return nil
}

func parseCommaDelimetedList(f string) []string {
	var res []string
	if f == "" {
		return res
	}
	for _, s := range strings.Split(f, ",") {
		s = strings.Trim(s, " ")
		res = append(res, s)
	}
	return res
}

func GetRepo() (*pb.SourceRepository, error) {
	var res *pb.SourceRepository
	var err error
	c := authremote.Context()
	if *repoid != 0 {
		res, err = pb.GetGIT2Client().RepoByID(c, &pb.ByIDRequest{ID: uint64(*repoid)})
	} else if *repourl != "" {
		res, err = pb.GetGIT2Client().RepoByURL(c, &pb.ByURLRequest{URL: *repourl})
	} else {
		err = fmt.Errorf("neither -repoid nor -repourl specified")
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func printDesc() error {
	sr, err := GetRepo()
	if err != nil {
		return err
	}
	s := strings.TrimSuffix(sr.Description, "\n") + "\n"
	fmt.Print(s)
	return nil
}
func denyMsg() error {
	sr, err := GetRepo()
	if err != nil {
		return err
	}
	req := &pb.DenyMessageRequest{
		RepositoryID: sr.ID,
		DenyMessage:  *message,
	}
	ctx := authremote.Context()
	_, err = pb.GetGIT2Client().SetDenyMessage(ctx, req)
	if err != nil {
		return err
	}
	fmt.Printf("Set Denymessage of repository #%d to \"%s\"\n", req.RepositoryID, req.DenyMessage)
	return nil
}

func GetTag() error {
	tag := *get_tag
	nreq := &pb.GetNTagRequest{RepositoryID: uint64(*repoid), Tag: tag}
	ctx := authremote.Context()
	nres, err := pb.GetGIT2Client().GetBuildWithNTag(ctx, nreq)
	if err != nil {
		return err
	}
	if nres.NTag == nil {
		fmt.Printf("No build with tag %d in repo %d\n", nreq.Tag, nreq.RepositoryID)
		return nil
	}
	fmt.Printf("Build: %d\n", nres.NTag.BuildID)
	return nil
}
func SetTag() error {
	tag := *get_tag
	nreq := &pb.AttachNTagRequest{RepositoryID: uint64(*repoid), BuildID: uint64(*build), Tag: tag}
	ctx := authremote.Context()
	_, err := pb.GetGIT2Client().AttachNTagToBuild(ctx, nreq)
	if err != nil {
		return err
	}
	fmt.Printf("Attached tag %d in repo %d to build %d\n", nreq.Tag, nreq.RepositoryID, nreq.BuildID)
	return nil
}
func Format() int {
	f := *format
	if f == "human" {
		return FORMAT_HUMAN
	} else if f == "shell" {
		return FORMAT_SHELL
	} else {
		panic(fmt.Sprintf("\"%s\" is not a valid format", f))
	}
}
