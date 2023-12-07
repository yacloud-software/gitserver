package builder

/*
git-hook connects via tcp to this monstrosity

*/
import (
	"bufio"
	"flag"
	"fmt"
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
func (t *TCPConn) Write(s []byte) (int, error) {
	fmt.Print(string(s))
	t.conn.Write(s)
	return len(s), nil
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
	ctx, err := gt.GetContext()
	if err != nil {
		t.Printf("Failed to get context: %s\n", err)
		return
	}
	if !*run_scripts {
		t.Writeln("Builds disabled.\n")
	} else {
		err = external_builder(ctx, gt, t)
	}
	if err != nil {
		t.Writeln(fmt.Sprintf("Failed to build: %s", err))
	}
}

func (t *TCPConn) Printf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	t.conn.Write([]byte(msg))
	fmt.Print("[hook] " + msg)
}

