package main

import (
	"context"
	"flag"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/server"
	"golang.conradwood.net/go-easyops/sql"
	"golang.conradwood.net/go-easyops/utils"
	"google.golang.org/grpc"
)

var (
	grpc_port = flag.Int("grpc_port", 4109, "grpc-server for git credentials server")
)

func StartServer() {
	psql, err := sql.Open()
	gitcreddb = db.NewDBGitCredentials(psql)
	utils.Bail("failed to create tables", gitcreddb.CreateTable(context.Background()))
	cserver := &CServer{}
	sd := server.NewServerDef()
	sd.Port = *grpc_port
	sd.Register = server.Register(func(server *grpc.Server) error {
		gitpb.RegisterGITCredentialsServer(server, cserver)
		return nil
	})
	err = server.ServerStartup(sd)
	utils.Bail("failed to start git credentials grpc server", err)

}

type CServer struct {
}

func (c *CServer) GitInvoked(ctx context.Context, req *gitpb.GitCredentialsRequest) (*gitpb.GitCredentialsResponse, error) {
	fmt.Printf("--------------------------------------\n")
	u := auth.GetUser(ctx)
	s := auth.GetService(ctx)
	fmt.Printf("User   : %s\n", auth.Description(u))
	fmt.Printf("Service: %s\n", auth.Description(s))
	/*
		fmt.Printf("Environment:\n")
		for _, e := range req.Environment {
			fmt.Printf("   %s\n", e)
		}
		fmt.Printf("Parameters:\n")
		for _, e := range req.Args {
			fmt.Printf("   %s\n", e)
		}
	*/
	fmt.Printf("Stdin:\n")
	fmt.Printf("%s\n", req.Stdin)

	ro, err := DoAuth(ctx, req.Args, req.Stdin)
	if err != nil {
		fmt.Printf("authentication failed: %s\n", err)
		return nil, err
	}
	res := &gitpb.GitCredentialsResponse{
		Stdout: ro,
	}
	return res, nil
}
