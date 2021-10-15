package builder

import (
	"context"
	"flag"
	"golang.conradwood.net/coderunners"
	"golang.conradwood.net/coderunners/registry"
	//	"encoding/base64"
	"fmt"
	am "golang.conradwood.net/apis/auth"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/gitserver/git"
	"golang.conradwood.net/gitserver/git2"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/sql"
	"golang.conradwood.net/go-easyops/utils"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

/*
  ep.setEnvironmentVariable("BUILD_NUMBER", "" + ce.getBuildNumber());
        ep.setEnvironmentVariable("JOB_NAME", b.getProjectName());
        ep.setEnvironmentVariable("PROJECT_NAME", b.getProjectName());
        ep.setEnvironmentVariable("GIT_BRANCH", b.getBranch());
        ep.setEnvironmentVariable("BUILD_DIR", dir);
        ep.setEnvironmentVariable("GOPATH", gwc.getGoPath());
        ep.setEnvironmentVariable("OLDREV", b.getOldRev());
        ep.setEnvironmentVariable("NEWREV", b.getNewRev());
        ep.setEnvironmentVariable("AUTHOR", gpd.getAuthor());
        ep.setEnvironmentVariable("COMMIT_ID", gpd.getCommitid());
        ep.setEnvironmentVariable("COMMIT_DATE", gpd.getDate());
        ep.setEnvironmentVariable("REPOSITORY_ID", "" + Main.getResolver().getRepositoryId(b.getProjectName()));
*/

const (
	RULES_REJECT = 1 // reject if broken build
	RULES_DO     = 2 // warn only
)

var (
	run_scripts   = flag.Bool("run_scripts", true, "if false, no automatic builds and checks will be created")
	BUILD_SCRIPTS = map[string]string{
		"STANDARD_PROTOS": "protos-build.sh",
		"STANDARD_GO":     "go-build.sh",
		"KICAD":           "kicad-build.sh",
		"STANDARD_JAVA":   "java-build.sh",
	}
)

type Builder struct {
	deploydir  string // dir where gitserver is deployed
	trigger    *GitTrigger
	git        *git.Git
	buildrules *BuildRules
	db         *sql.DB
	buildv2    *gitpb.Build
	conn       *TCPConn
	repository *BuildRepo
	token      string
}
type BuildRules struct {
	Prebuild   int
	PostCommit int
	Builds     []string
	Targets    []string
}

func (b *BuildRules) TargetGoOS() []string {
	if b.Targets == nil || len(b.Targets) == 0 {
		return []string{"linux"}
	}
	return b.Targets
}
func (b *BuildRules) JavaBuild() bool {
	for _, b := range b.Builds {
		if b == "STANDARD_JAVA" {
			return true
		}
	}
	return false
}
func (b *BuildRules) ProtosBuild() bool {
	for _, b := range b.Builds {
		if b == "STANDARD_PROTOS" {
			return true
		}
	}
	return false
}

func (b *BuildRules) GoBuild() bool {
	for _, b := range b.Builds {
		if b == "STANDARD_GO" {
			return true
		}
	}
	return false
}

// return scriptname or ""
func (b *BuildRules) CheckBuildType(buildtype string) string {
	for _, b := range b.Builds {
		if b == buildtype {
			// build activated in BUILD_RULES
			sc := BUILD_SCRIPTS[b]
			return sc
		}
	}
	return ""
}
func (b *BuildRules) KicadBuild() bool {
	for _, b := range b.Builds {
		if b == "KICAD" {
			return true
		}
	}
	return false
}
func (b *BuildRules) AutoBuild() bool {
	for _, b := range b.Builds {
		if b == "AUTOBUILD_SH" {
			return true
		}
	}
	return false
}

func parseAction(s string) (int, error) {
	if s == "reject" {
		return RULES_REJECT, nil
	} else if s == "do" {
		return RULES_DO, nil
	} else {
		return -1, fmt.Errorf("buildrules: Tag %s unknown", s)
	}
}

func (b *Builder) readBuildrules() error {
	rules := &BuildRules{}
	b.buildrules = rules
	br := b.git.FullDirectoryPath() + "/BUILD_RULES"
	if !utils.FileExists(br) {
		return nil
	}
	fc, err := ioutil.ReadFile(br)
	if err != nil {
		return err
	}
	gotBuilds := false
	for ln, line := range strings.Split(string(fc), "\n") {
		if len(line) < 2 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		sp := strings.SplitN(line, "=", 2)
		if len(sp) < 1 {
			return fmt.Errorf("buildrules: Line %d invalid (only %d parts) [%s]", ln+1, len(sp), line)
		}
		if sp[0] == "PREBUILD" {
			rules.Prebuild, err = parseAction(sp[1])
		} else if sp[0] == "POSTCOMMIT" {
			rules.PostCommit, err = parseAction(sp[1])
		} else if sp[0] == "BUILDS" {
			gotBuilds = true
			for _, bs := range strings.Split(sp[1], ",") {
				rules.Builds = append(rules.Builds, bs)
			}
		} else if sp[0] == "TARGETS" {
			for _, bs := range strings.Split(sp[1], ",") {
				rules.Targets = append(rules.Targets, bs)
			}
		} else {
			return fmt.Errorf("buildrules: Line %d invalid (invalid tag \"%s\") [%s]", ln+1, sp[0], line)
		}
		if err != nil {
			return err
		}
	}

	// set default to autobuild.sh if it exists
	if !gotBuilds {
		if utils.FileExists(b.git.FullDirectoryPath() + "/autobuild.sh") {
			rules.Builds = []string{"AUTOBUILD_SH"}
		}
	}
	return nil
}
func (b *Builder) BuildGit(ctx context.Context) error {
	var err error
	b.db, err = sql.Open()
	if err != nil {
		return err
	}
	b.repository, err = b.createRepositoryFromHook()
	if err != nil {
		return err
	}
	fmt.Printf("Repository ID #%d\n", b.repository.GetID())

	err = b.readBuildrules()
	if err != nil {
		fmt.Printf("Failed to read buildrules: %s\n", err)
		return err
	}
	// special ones:
	err = b.CreateBuild()
	if err != nil {
		fmt.Printf("Failed to create new build: %s\n", utils.ErrorString(err))
		return err
	}
	b.conn.Printf("BuildID #%d (Build-Version %d)\n", b.Build().ID, b.trigger.gitinfo.Version)

	fmt.Printf("Configured/Detected Build types: %s\n", b.buildrules.Builds)
	target_os := strings.Join(b.buildrules.TargetGoOS(), " ")
	if !*run_scripts {
		return nil
	}
	err = b.buildscript(ctx, script("clean-build.sh"), target_os)
	if err != nil {
		return err
	}
	o := &registry.Opts{
		BuildID:        b.Build().ID,
		BuildTimestamp: uint32(time.Now().Unix()),
		RepositoryID:   b.repository.GetID(),
		ArtefactName:   b.repository.ArtefactName(),
		CommitUserID:   b.trigger.gitinfo.User.ID,
	}
	err = coderunners.RunGit(o)
	if err != nil {
		return err
	}
	if b.buildrules.AutoBuild() { // run this first, it might destroy previous builds
		err = b.buildscript(ctx, "./autobuild.sh", "all")
		if err != nil {
			return err
		}
	}

	for rulename, _ := range BUILD_SCRIPTS {
		scriptname := b.buildrules.CheckBuildType(rulename)
		if scriptname == "" {
			continue
		}
		err = b.buildscript(ctx, script(scriptname), target_os)
		if err != nil {
			return err
		}
	}

	// ** now create the dist and upload it
	err = b.buildscript(ctx, script("dist.sh"), target_os)
	if err != nil {
		return err
	}
	// ** now update the database with our information ****
	err = b.updateBuild()
	if err != nil {
		return err
	}
	// Done
	b.conn.Printf("\n\n*************** BUILD #%d (repo %d) SUCCESSFUL *******\n\n", b.Build().ID, b.Build().RepositoryID)
	return nil
}

// run autobuild.sh...
func (b *Builder) buildscript(ctx context.Context, fscript string, targets string) error {
	fq := b.git.FullDirectoryPath() + "/" + fscript
	if !utils.FileExists(fscript) && !utils.FileExists(fq) {
		return fmt.Errorf("%s nor %s does exist", fscript, fq)
	}
	cmd := exec.Command(fscript)
	cmd.Dir = b.git.FullDirectoryPath()
	fmt.Printf("Executing script %s in cwd \"%s\"\n", fscript, cmd.Dir)
	if cmd.Dir == "" || !utils.FileExists(cmd.Dir) {
		return fmt.Errorf("Directory \"%s\" does not exist\n", cmd.Dir)
	}
	ep, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	op, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go b.pipeOutput(ep)
	go b.pipeOutput(op)
	cmd.Env = b.env()
	b.addContextEnv(ctx, cmd)
	cmd.Env = append(cmd.Env, fmt.Sprintf("BUILD_TARGETS=%s", targets))
	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
func (b *Builder) pipeOutput(rc io.ReadCloser) {
	buf := make([]byte, 1024)
	for {
		n, err := rc.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Failed to read %s:\n", err)
			break
		}
		b.conn.conn.Write(buf[:n])
		fmt.Printf("%s", string(buf[:n]))
	}

}

func (b *Builder) env() []string {
	// standard environment variables...
	std := `
JAVA_HOME=/etc/java-home
GRADLE_HOME=/srv/java/gradle/latest
TERM=xterm
SHELL=/bin/bash
ANT_HOME=/srv/java/ant/current/
PATH=/opt/cnw/ctools/dev/bin:/opt/cnw/ctools/dev/go/current/go/bin/:/etc/java-home/bin:/srv/singingcat/binutils/bin/:~/bin:/sbin:/usr/sbin:/usr/local/bin:/usr/bin:/bin:/srv/java/ant/current/bin:/srv/singingcat/esp8266/sdk/xtensa-lx106-elf/bin/:/srv/java/ant/bin:/srv/java/gradle/latest/bin
PWD=/tmp
GOROOT=/opt/cnw/ctools/dev/go/current/go
LANG=en_GB.UTF-8
LANGUAGE=en_GB:en
LC_CTYPE=en_GB.UTF-8
`
	var res []string
	for _, s := range strings.Split(std, "\n") {
		if len(s) < 2 {
			continue
		}
		res = append(res, s)
	}
	dir, err := os.Getwd()
	bindir := "./"
	if err != nil {
		fmt.Printf("Unable to get current directory. (%s)\n", err)
	} else {
		bindir = dir
	}
	absdir, err := filepath.Abs(b.git.FullDirectoryPath())
	if err != nil {
		fmt.Printf("Cannot get absolute directory: %s\n", err)
		return nil
	}
	fmt.Printf("Bindir: \"%s\"\n", bindir)
	os.MkdirAll(bindir+"/gobin", 0777)
	os.MkdirAll(bindir+"/gocache", 0777)
	os.MkdirAll(bindir+"/gotmp", 0777)
	res = append(res, fmt.Sprintf("BUILD_NUMBER=%d", b.Build().ID))
	res = append(res, fmt.Sprintf("GOPATH=%s", absdir))
	res = append(res, fmt.Sprintf("HOME=%s", absdir))
	res = append(res, fmt.Sprintf("BUILD_DIR=%s", absdir))
	res = append(res, fmt.Sprintf("COMMIT_ID=%s", b.trigger.newrev))
	res = append(res, fmt.Sprintf("REPOSITORY_ID=%d", b.repository.GetID()))
	res = append(res, fmt.Sprintf("PROJECT_NAME=%s", b.repository.RepoName()))
	res = append(res, fmt.Sprintf("BUILD_REPOSITORY=%s", b.repository.RepoName()))
	res = append(res, fmt.Sprintf("BUILD_ARTEFACT=%s", b.repository.ArtefactName()))
	res = append(res, fmt.Sprintf("BUILD_TIMESTAMP=%d", b.Build().Timestamp))
	res = append(res, fmt.Sprintf("GIT_BRANCH=%s", "master"))
	res = append(res, fmt.Sprintf("GOBIN=%s/gobin", bindir))
	res = append(res, fmt.Sprintf("GOCACHE=%s/gocache", bindir))
	res = append(res, fmt.Sprintf("GOTMPDIR=%s/gotmp", bindir))
	res = append(res, fmt.Sprintf("SCRIPTDIR=%s", scriptsdir))

	return res
}
func (b *Builder) addContextEnv(ctx context.Context, cmd *exec.Cmd) error {
	u := auth.GetUser(ctx)
	if u == nil {
		fmt.Printf("WARNING: no user in context\n")
	} else {
		fmt.Printf("Executing scripts as user %s (%s)\n", u.ID, auth.Description(u))
	}
	ncb, err := auth.SerialiseContextToString(ctx)
	if err != nil {
		fmt.Printf("Failed to encode context to string: %s\n", err)
		return err
	}

	ncs := fmt.Sprintf("GE_CTX=%s", ncb)
	for i, e := range cmd.Env {
		if strings.HasPrefix(e, "GE_CTX=") {
			cmd.Env[i] = ncs
			return nil
		}
	}
	cmd.Env = append(cmd.Env, ncs)
	if u != nil {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GE_USER_EMAIL=%s", u.Email))
		cmd.Env = append(cmd.Env, fmt.Sprintf("GE_USER_ID=%s", u.ID))
		if b.token == "" {
			tr, err := GetAuthManagerClient().GetTokenForMe(ctx, &am.GetTokenRequest{DurationSecs: 300})
			if err != nil {
				fmt.Printf("unable to get authentication token for external script(s): %s\n", utils.ErrorString(err))
				return err
			}
			b.token = tr.Token
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("GE_USER_TOKEN=%s", b.token))
	}
	return nil
}
func GetAuthManagerClient() am.AuthManagerServiceClient {
	return authremote.GetAuthManagerClient()
}

/***********************************************************************************
* buildrepo
************************************************************************************/
// build repo is a somewhat hacky abstraction of v1/v2 repositories
type BuildRepo struct {
	id           uint64
	reponame     string
	artefactname string
}

func (b *Builder) createRepositoryFromHook() (*BuildRepo, error) {
	ctx, err := b.trigger.GetContext()
	if err != nil {
		fmt.Printf("Failed to get context: %s\n", err)
		return nil, err
	}
	if b.trigger.gitinfo.Version == 2 {
		rid := b.trigger.gitinfo.RepositoryID
		sr, err := gitpb.GetGIT2Client().RepoByID(ctx, &gitpb.ByIDRequest{ID: rid})
		if err != nil {
			fmt.Printf("Failed to get V2 repo (%d) by id: %s\n", rid, err)
			return nil, err
		}

		res := &BuildRepo{
			id:           sr.ID,
			reponame:     fmt.Sprintf("repo-%d", sr.ID), // V2 repos have no canonical name
			artefactname: sr.ArtefactName,
		}
		return res, nil
	}
	/*
		// version "1" old style
		repoinfo, err := gs.RepoByID(ctx, &gitpb.ByIDRequest{ID: b.trigger.gitinfo.RepositoryID})
		if err != nil {
			fmt.Printf("Failed to get repo by id: %s\n", err)
			return nil, err
		}

		res := &BuildRepo{
			id:           repoinfo.Repository.ID,
			reponame:     repoinfo.Repository.RepoName,
			artefactname: repoinfo.Repository.ArtefactName,
		}
		return res, nil
	*/
	return nil, fmt.Errorf("Version %d not implemented in git-build.go", b.trigger.gitinfo.Version)
}

func (b *BuildRepo) GetID() uint64 {
	return b.id
}
func (b *BuildRepo) RepoName() string {
	return b.reponame
}
func (b *BuildRepo) ArtefactName() string {
	return b.artefactname
}

/**************************************************
create a new build
/*************************************************/
func (b *Builder) Build() *gitpb.Build {
	return b.buildv2
}

// creates new builds, v2
func (b *Builder) CreateBuild() error {
	// create v2
	bdb := db.NewDBBuild(git2.GetDatabase())
	nb := &gitpb.Build{
		UserID:       b.trigger.UserID(),
		RepositoryID: b.repository.GetID(),
		CommitHash:   b.trigger.newrev,
		Branch:       b.trigger.Branch(),
		LogMessage:   b.git.LogMessage,
		Timestamp:    uint32(time.Now().Unix()),
	}

	id, err := bdb.Save(context.Background(), nb)
	if err != nil {
		fmt.Printf("Failed to save to database: %s\n", err)
		return err
	}
	b.buildv2 = nb
	b.buildv2.ID = id

	return nil
}

// updates both builds, v1 and v2
func (b *Builder) updateBuild() error {
	//v2
	bdb := db.NewDBBuild(git2.GetDatabase())
	b.buildv2.Success = true
	err := bdb.Update(context.Background(), b.buildv2)
	if err != nil {
		return err
	}

	return nil
}
