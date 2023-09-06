package main

import (
	"context"
	"fmt"
	"golang.conradwood.net/gitserver/db"
	"time"
)

func cleaner_loop() {
	t := time.Duration(3) * time.Second
	for {
		time.Sleep(t)
		t = time.Duration(30) * time.Second
		err := clean_creds()
		if err != nil {
			fmt.Printf("failed to clean creds: %s\n", err)
		}

	}
}
func clean_creds() error {
	remotes, err := db.DefaultDBInternalGitHost().All(context.Background())
	if err != nil {
		return err
	}
	exp := uint32(time.Now().Unix())
	for _, r := range remotes {
		if r.Expiry <= exp {
			ctx := context.Background()
			err := db.DefaultDBInternalGitHost().DeleteByID(ctx, r.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil

}
