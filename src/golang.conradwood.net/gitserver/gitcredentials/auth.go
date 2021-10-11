package main

import (
	"context"
	"fmt"
	pbauth "golang.conradwood.net/apis/auth"
	pb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/gitserver/query"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
	"strings"
)

var (
	gitcreddb *db.DBGitCredentials
)

func DoAuth(ctx context.Context, args []string, stdin string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("missing arg")
	}
	if args[0] == "get" {
		return GetAuth(ctx, stdin)
	}
	fmt.Printf("Unhandled operation \"%s\"\n", args[0])
	return "", nil

}
func GetAuth(ctx context.Context, stdin string) (string, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return "", errors.Unauthenticated(ctx, "need auth")
	}
	keys, err := parseStdin(stdin)
	if err != nil {
		return "", err
	}
	hostname := keys["hostname"]
	if hostname == "" {
		hostname = keys["host"]
	}
	keys["hostname"] = hostname
	if hostname == "" {
		return "", errors.InvalidArgs(ctx, "missing hostname", "missing hostname")
	}

	fmt.Printf("User %s requests auth for host \"%s\"\n", u.ID, hostname)
	_, ourhost, err := query.SendPing(ctx, hostname)
	if err != nil {
		fmt.Printf("Ping failed (%s)\n", err)
		ourhost = false
	}
	if !ourhost {
		// do not send tokens to hosts other than us
		return remoteHost(ctx, keys)
	}

	at := &pbauth.GetTokenRequest{DurationSecs: 60}
	as, err := authremote.GetAuthManagerClient().GetTokenForMe(ctx, at)
	if err != nil {
		return "", err
	}
	uname := fmt.Sprintf("%s@token.yacloud.eu", u.ID)
	password := as.Token
	res := fmt.Sprintf("username=%s\npassword=%s\n", uname, password)
	fmt.Printf("Authenticating host \"%s\" as internal, user %s\n", hostname, uname)
	return res, nil
}
func parseStdin(stdin string) (map[string]string, error) {
	res := make(map[string]string)
	for _, l := range strings.Split(stdin, "\n") {
		ts := strings.SplitN(l, "=", 2)
		if len(ts) != 2 {
			continue
		}
		res[ts[0]] = ts[1]
	}
	return res, nil
}

func remoteHost(ctx context.Context, keys map[string]string) (string, error) {
	hostname := keys["hostname"]
	fmt.Printf("Host \"%s\" is not an internal host\n", hostname)
	u := auth.GetUser(ctx)
	userid := u.ID
	creds, err := gitcreddb.ByUserID(ctx, userid)
	if err != nil {
		fmt.Printf("No credentials for user \"%s\"\n", userid)
		return "", err
	}
	var use_creds *pb.GitCredentials
	for _, c := range creds {
		if c.Host == hostname {
			use_creds = c
			break
		}
	}
	if use_creds == nil {
		fmt.Printf("No credentials for user \"%s\" to host \"%s\"\n", userid, hostname)
		return fmt.Sprintf("quit=1\n"), nil
	}
	res := fmt.Sprintf("username=%s\npassword=%s\n", use_creds.Username, use_creds.Password)
	return res, nil

}
