package git2

import (
	"context"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
	"net/http"
)

// send a HTTPResponse server to an http writer
func NewHTTPWriter(h *HTTPRequest, ctx context.Context) *http_writer {
	res := &http_writer{ctx: ctx, h: h}
	return res
}

type http_writer struct {
	h   *HTTPRequest
	ctx context.Context
}

func (hw *http_writer) Context() context.Context {
	return hw.ctx
}
func (hw *http_writer) Send(hr *gitpb.HookResponse) error {
	if hr.ErrorMessage != "" {
		hw.h.w.Write([]byte(hr.ErrorMessage))
		return fmt.Errorf("%s", hr.ErrorMessage)
	}
	b := []byte(hr.Output)
	n, err := hw.h.w.Write(b)
	if len(b) != n {
		return fmt.Errorf("short write (%d != %d)", n, len(b))
	}
	if f, ok := hw.h.w.(http.Flusher); ok {
		f.Flush()
	}
	return err
}

