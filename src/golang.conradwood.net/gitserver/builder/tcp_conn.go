package builder

import (
	"bufio"
	"flag"
	"fmt"
	"golang.conradwood.net/gitserver/git"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/utils"
	"net"
)

var (
	run_scripts = flag.Bool("run_scripts", true, "if false, no automatic builds and checks will be created")
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
	err = t.build(gt)
	if err != nil {
		t.Writeln(fmt.Sprintf("Failed to build: %s", err))
	}
}

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
