package main

import (
	"fmt"
	pb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
)

func DoBuilds() {
	ctx := authremote.Context()
	sr := GetRepository(ctx)
	fmt.Printf("Repository  : #%d\n", sr.ID)
	lreq := &pb.ByIDRequest{ID: sr.ID}
	bl, err := pb.GetGIT2Client().GetRecentBuilds(ctx, lreq)
	utils.Bail("failed to get builds", err)
	t := &utils.Table{}
	for _, b := range bl.Builds {
		t.AddUint64(b.ID)
		t.AddString(utils.TimestampString(b.Timestamp))
		t.AddString(b.CommitHash)
		t.NewRow()
	}
	fmt.Println(t.ToPrettyString())
}
