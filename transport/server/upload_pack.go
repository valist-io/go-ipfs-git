package server

import (
	"context"
	"io"

	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/sideband"
	"github.com/go-git/go-git/v5/plumbing/revlist"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

const packWindowSize = 10

type uploadPackSession struct {
	session
}

func (s *uploadPackSession) UploadPack(ctx context.Context, req *packp.UploadPackRequest) (*packp.UploadPackResponse, error) {
	if req.IsEmpty() {
		return nil, transport.ErrEmptyUploadPackRequest
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	haves, err := revlist.Objects(s.storer, req.Haves, nil)
	if err != nil {
		return nil, err
	}

	objs, err := revlist.Objects(s.storer, req.Wants, haves)
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	var w io.Writer = pw

	switch {
	case req.Capabilities.Supports(capability.Sideband):
		w = sideband.NewMuxer(sideband.Sideband, pw)
	case req.Capabilities.Supports(capability.Sideband64k):
		w = sideband.NewMuxer(sideband.Sideband64k, pw)
	}

	e := packfile.NewEncoder(w, s.storer, false)
	go func() {
		_, err := e.Encode(objs, packWindowSize)
		pw.CloseWithError(err)
	}()

	return packp.NewUploadPackResponseWithPackfile(req, pr), nil
}
