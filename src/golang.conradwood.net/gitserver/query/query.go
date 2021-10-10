package query

import (
	"context"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/http"
	"golang.conradwood.net/go-easyops/sql"
	"golang.conradwood.net/go-easyops/utils"
	"time"
)

var (
	psql *sql.DB
)

func startup() error {
	if psql != nil {
		return nil
	}
	var err error
	psql, err = sql.Open()
	if err != nil {
		return err
	}
	return nil
}

type PingResponder interface {
	Error(error)
	Write([]byte)
	SingleHeader(string) string
}

// we received a ping. respond to it
func Ping(h PingResponder) {
	err := startup()
	if err != nil {
		h.Error(fmt.Errorf("internal error"))
		return
	}
	fmt.Printf("Responding to ping...\n")
	actok := h.SingleHeader("x-associationtoken")
	if actok == "" {
		h.Error(fmt.Errorf("invalid unauthenticated request"))
		return
	}
	ctx := authremote.Context()
	dbs, err := db.NewDBPingState(psql).ByAssociationToken(ctx, actok)
	if err != nil {
		h.Error(err)
		return
	}
	if len(dbs) == 0 {
		h.Error(fmt.Errorf("invalid associationtoken"))
		return
	}
	ps := dbs[0]
	s, err := utils.MarshalBytes(ps)
	if err != nil {
		h.Error(fmt.Errorf("unable to marshal pingstate"))
		return
	}
	h.Write(s)
}

// send a ping
func SendPing(ctx context.Context, host string) (*gitpb.PingState, bool, error) {
	err := startup()
	if err != nil {
		return nil, false, err
	}
	cutoff := uint32(time.Now().Unix()) - 30 // valid for 30 seconds
	_, err = psql.ExecContext(ctx, "delete_old_ping_state", "delete from pingstate where created < $1", cutoff)
	if err != nil {
		fmt.Printf("Failed to delete old pings: %s\n", utils.ErrorString(err))
	}
	ps := &gitpb.PingState{
		AssociationToken: utils.RandomString(256),
		Created:          uint32(time.Now().Unix()),
		ResponseToken:    utils.RandomString(256),
	}
	_, err = db.NewDBPingState(psql).Save(ctx, ps)
	if err != nil {
		return nil, false, err
	}
	h := http.HTTP{}
	url := fmt.Sprintf("https://%s/git/self/Ping", host)
	h.SetHeader("X-AssociationToken", ps.AssociationToken)
	hb := h.Get(url)
	err = hb.Error()
	if err != nil {
		// it is not really an error if remote doesn't have this url
		if hb.HTTPCode() == 404 {
			return nil, false, nil
		}
		fmt.Printf("Ping code: %d\n", hb.HTTPCode())
		return nil, false, err
	}
	psr := &gitpb.PingState{}
	err = utils.UnmarshalBytes(hb.Body(), psr)
	if err != nil {
		return nil, false, err
	}

	v := ps.ResponseToken == psr.ResponseToken
	if !v {
		fmt.Printf("Token 1: %s\n", psr.ResponseToken)
		fmt.Printf("Token 2: %s\n", ps.ResponseToken)
	}
	fmt.Printf("Returned pingstate id: %d, created id: %d (token match: %v)\n", psr.ID, ps.ID, v)
	if psr.ID == ps.ID && ps.ResponseToken == psr.ResponseToken {
		return ps, true, nil
	} else {
		return ps, false, nil
	}
}
