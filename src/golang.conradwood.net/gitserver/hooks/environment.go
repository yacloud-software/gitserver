package main

import (
	"context"
	"fmt"
	apb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/ctx"
	"golang.conradwood.net/go-easyops/tokens"
	//	"golang.conradwood.net/go-easyops/utils"
	"os"
)

type Environment struct {
	ctx         context.Context
	CurrentUser *apb.User
}

func Setup() *Environment {
	tokens.DisableUserToken()
	res := &Environment{}

	cs := os.Getenv("GE_CTX")
	res.ctx = parseContextFromEnv(cs)
	res.CurrentUser = auth.GetUser(res.ctx)
	fmt.Printf("[hook] Current user: %s\n", auth.CurrentUserString(res.ctx))
	if res.CurrentUser == nil {
		fmt.Printf("[hook] Environment Variable GE_CTX:\n\"%s\"\n", os.Getenv("GE_CTX"))
	}
	return res
}
func parseContextFromEnv(env string) context.Context {
	if env == "" {
		fmt.Printf("[hook] WARNING no serialised context in git hook\n")
		return authremote.Context()
	}

	var c context.Context
	var err error
	b := []byte(env)
	//fmt.Printf("[hook] Context: %s (%s)\n", utils.HexStr(b[:15]), string(b[:15]))
	if ctx.IsSerialisedByBuilder(b) {
		c, err = ctx.DeserialiseContextFromString(env)
		if err == nil {
			return c
		}
		fmt.Printf("[hook] failed to deserialise builder context (%s)\n", err)
	}
	c = authremote.Context() // try old way
	return c

}

func (e *Environment) IsRoot() bool {
	return auth.IsRoot(e.ctx)
}
func (e *Environment) IsYacloudAdmin() bool {
	// in group "8" (yacloud-admin) ?
	return auth.IsInGroup(e.ctx, "8")
}



