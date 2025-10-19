package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/client"
)

func call_gitserver(req *gitpb.HookRequest) error {
	e := Setup()
	req.Args = flag.Args()
	req.Environment = os.Environ()
	ctx := e.ctx
	gip := "localhost:" + os.Getenv("GITSERVER_GRPC_PORT")
	fmt.Printf("grpc connection to \"%s\"\n", gip)
	con, err := client.ConnectWithIP(gip)
	if err != nil {
		return err
	}
	defer con.Close()
	gc := gitpb.NewGIT2Client(con)
	srv, err := gc.RunLocalHook(ctx, req)
	if err != nil {
		return err
	}
	for {
		hr, err := srv.Recv()
		if hr != nil {
			if hr.Output != "" {
				fmt.Print(hr.Output)
			}
			if hr.ErrorMessage != "" {
				return fmt.Errorf("%s", hr.ErrorMessage)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}
