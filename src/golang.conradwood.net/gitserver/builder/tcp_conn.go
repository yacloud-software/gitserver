package builder

import (
	"bufio"
	"flag"
	"fmt"
	"golang.conradwood.net/apis/gitbuilder"
	"golang.conradwood.net/gitserver/git"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/utils"
	"io"
	"net"
)

var (
	use_external_builder = flag.Bool("use_external_builder", false, "if true, the external builder will be used")
	run_scripts          = flag.Bool("run_scripts", true, "if false, no automatic builds and checks will be created")
)

type TCPConn struct {
	conn net.Conn
}

func NewTCPConn(c net.Conn) *TCPConn {
	res := &TCPConn{conn: c}
	return res
}
func (t *TCPConn) Writeln(s string) {
	fmt.Println(s)
	t.conn.Write([]byte(s + "\n"))
}
func (t *TCPConn) Write(s []byte) {
	fmt.Print(string(s))
	t.conn.Write(s)
}
func (t *TCPConn) HandleConnection() {
	defer t.conn.Close()
	fmt.Printf("received git request\n")
	t.Writeln("building git repository...")
	r := bufio.NewReader(t.conn)
	line, err := r.ReadString('\n')
	if err != nil {
		t.Writeln(fmt.Sprintf("Failed to read header: %s", err))
		return
	}
	gt, err := ParseGitTrigger(line)
	if err != nil {
		t.Printf("Failed to parse git trigger: %s\n", utils.ErrorString(err))
		return
	}
	if *use_external_builder {
		err = t.external_builder(gt)
	} else {
		err = t.build(gt)
	}
	if err != nil {
		t.Writeln(fmt.Sprintf("Failed to build: %s", err))
	}
}

// use the gitbuilder service to build instead of locally forking it off
// e.g. ref== 'master', newrev == commitid
func (t *TCPConn) external_builder(gt *GitTrigger) error {
	ctx, err := gt.GetContext()
	if err != nil {
		return err
	}
	gb := gitbuilder.GetGitBuilderClient()
	br := &gitbuilder.BuildRequest{
		GitURL:   gt.gitinfo.URL,
		CommitID: gt.newrev,
	}
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
			t.Write(res.Stdout)
		}
	}
	if !lastResponse.Success {
		return fmt.Errorf("build failed (%s)", lastResponse.ResultMessage)
	}
	return nil
}

// run scripts directly within git-server host (forking off scripts)
// e.g. ref== 'master', newrev == commitid
func (t *TCPConn) build(gt *GitTrigger) error {

	if !*run_scripts {
		t.Writeln("Builds disabled.\n")
		return nil
	}
	t.Writeln(fmt.Sprintf("Building %s, commit %s on %s", gt.ref, gt.repodir, gt.newrev))
	ctx, err := gt.GetContext()
	if err != nil {
		fmt.Printf("No context for user \"%s\": %s\n", gt.UserID(), err)
		//		return err
	} else {
		u := auth.GetUser(ctx)
		if u != nil {
			s := fmt.Sprintf("Executing build as User %s (%s)", u.ID, auth.Description(u))
			t.Writeln(s)
		}
	}
	g := git.GitRepo(gt.repodir)
	defer g.Close()
	t.Writeln(fmt.Sprintf("Temporary Working ID: %d in %s\n", g.ID, g.FullDirectoryPath()))
	err = g.Checkout(ctx, gt.ref, gt.newrev)
	if err != nil {
		return err
	}
	b := &Builder{git: g, conn: t, trigger: gt}
	err = b.BuildGit(ctx)
	if err != nil {
		fmt.Printf("Failed to build: %s\n", err)
	}
	return err
}

func (t *TCPConn) Printf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	t.conn.Write([]byte(msg))
	fmt.Print("[hook] " + msg)
}
