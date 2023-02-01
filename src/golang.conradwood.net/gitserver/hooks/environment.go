package main

import (
	"context"
	"fmt"
	apb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/ctx"
	"golang.conradwood.net/go-easyops/tokens"
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
	if cs != "" {
		var c context.Context
		var err error
		if ctx.IsSerialisedByBuilder([]byte(cs)) {
			c, err = ctx.DeserialiseContextFromString(cs)
			if err != nil {
				fmt.Printf("Failed to deserialise context (will try 'old' method): %s\n", err)
				c = authremote.Context() // try old way
			}
		} else {
			c = authremote.Context() // try old way
		}
		res.ctx = c
	} else {
		res.ctx = authremote.Context()
	}

	res.CurrentUser = auth.GetUser(res.ctx)
	if res.CurrentUser == nil {
		fmt.Printf("Environment Variable GE_CTX:\n\"%s\"\n", os.Getenv("GE_CTX"))
	}
	fmt.Printf("Current user: %s\n", auth.CurrentUserString(res.ctx))
	return res
}

func (e *Environment) IsRoot() bool {
	return auth.IsRoot(e.ctx)
}
func (e *Environment) IsYacloudAdmin() bool {
	// in group "8" (yacloud-admin) ?
	return auth.IsInGroup(e.ctx, "8")
}
