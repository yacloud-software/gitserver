package main

import (
	"bufio"
	"flag"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/cmdline"
	"golang.conradwood.net/go-easyops/utils"
	"io"
	"os"
)

func main() {
	flag.Parse()
	a := &gitpb.GitCredentialsRequest{
		Args:        flag.Args(),
		Environment: os.Environ(),
	}
	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, 8192)
	var total []byte
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			z := buf[:n]
			total = append(total, z...)
		}
		if err == io.EOF {
			break
		}
		utils.Bail("failed to read from stdin", err)
	}
	a.Stdin = string(total)
	ctx := authremote.Context()
	if ctx == nil || auth.GetUser(ctx) == nil {
		s := cmdline.GetEnvContext()
		if len(s) > 10 {
			s = s[:10]
		}
		b := []byte(s)
		fmt.Fprintf(os.Stderr, "Unable to create user context! (%s %s)\n", utils.HexStr(b), string(b))
		utils.WriteFile("/tmp/context.env", []byte(cmdline.GetEnvContext()))
		os.Exit(10)
	}
	r, err := gitpb.GetGITCredentialsClient().GitInvoked(ctx, a)
	utils.Bail("failed to call gitcredentials server", err)
	fmt.Printf("%s", r.Stdout)
	os.Exit(0)

}

