package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/sideband"
	"github.com/go-git/go-git/v5/plumbing/transport"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"

	"github.com/valist-io/go-ipfs-git/storage"
	"github.com/valist-io/go-ipfs-git/transport/server"
)

const (
	receivePackPath    = "/git-receive-pack"
	uploadPackPath     = "/git-upload-pack"
	advertisedRefsPath = "/info/refs"
)

type handler struct {
	api coreiface.CoreAPI
}

func ListenAndServe(api coreiface.CoreAPI, addr string) error {
	return http.ListenAndServe(addr, &handler{api})
}

func (s *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// TODO write errors to pkt line
	switch {
	case strings.HasSuffix(req.URL.Path, uploadPackPath):
		s.uploadPack(w, req)
	case strings.HasSuffix(req.URL.Path, receivePackPath):
		s.receivePack(w, req)
	case strings.HasSuffix(req.URL.Path, advertisedRefsPath):
		s.advertisedReferences(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (s *handler) loadStorage(ctx context.Context, rpath string) (*storage.Storage, error) {
	rpath = strings.TrimPrefix(rpath, "/")
	if rpath == "" {
		return storage.NewStorage(ctx, s.api.Dag())
	}

	res, err := s.api.ResolvePath(ctx, path.New(rpath))
	if err != nil {
		return nil, err
	}

	return storage.LoadStorage(ctx, s.api.Dag(), res.Cid())
}

func (s *handler) uploadPack(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	raw := strings.TrimSuffix(req.URL.Path, uploadPackPath)

	storer, err := s.loadStorage(ctx, raw)
	if err != nil {
		return err
	}

	sessreq := packp.NewUploadPackRequest()
	if err := sessreq.Decode(req.Body); err != nil {
		return err
	}

	sess := server.NewUploadPackSession(storer)
	sessres, err := sess.UploadPack(ctx, sessreq)
	if err != nil {
		return err
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/x-git-upload-pack-result")
	w.WriteHeader(http.StatusOK)

	if sessres != nil {
		sessres.Encode(w)
	}

	// incase sideband muxer didn't send final pkt line
	fmt.Fprintf(w, "0000")
	return nil
}

func (s *handler) receivePack(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	raw := strings.TrimSuffix(req.URL.Path, receivePackPath)

	storer, err := s.loadStorage(ctx, raw)
	if err != nil {
		return err
	}

	sessreq := packp.NewReferenceUpdateRequest()
	if err := sessreq.Decode(req.Body); err != nil {
		return err
	}

	sess := server.NewReceivePackSession(storer)
	sessres, err := sess.ReceivePack(ctx, sessreq)
	if err != nil {
		return err
	}

	node, err := storer.Node()
	if err != nil {
		return err
	}

	if err := s.api.Dag().Pinning().Add(ctx, node); err != nil {
		return err
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/x-git-receive-pack-result")
	w.WriteHeader(http.StatusOK)

	var out io.Writer = w
	switch {
	case sessreq.Capabilities.Supports(capability.Sideband):
		out = sideband.NewMuxer(sideband.Sideband, w)
	case sessreq.Capabilities.Supports(capability.Sideband64k):
		out = sideband.NewMuxer(sideband.Sideband64k, w)
	}

	if sessres != nil {
		sessres.Encode(out)
	}

	if mux, ok := out.(*sideband.Muxer); ok {
		mux.WriteChannel(sideband.ProgressMessage, []byte("\nRepository pinned successfully:\n\t"))
		mux.WriteChannel(sideband.ProgressMessage, []byte(node.Cid().String()))
	}

	// incase sideband muxer didn't send final pkt line
	fmt.Fprintf(w, "0000")
	return nil
}

func (s *handler) advertisedReferences(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	svc := req.URL.Query().Get("service")
	raw := strings.TrimSuffix(req.URL.Path, advertisedRefsPath)

	storer, err := s.loadStorage(ctx, raw)
	if err != nil {
		return err
	}

	var sess transport.Session
	switch svc {
	case transport.UploadPackServiceName:
		sess = server.NewUploadPackSession(storer)
	case transport.ReceivePackServiceName:
		sess = server.NewReceivePackSession(storer)
	default:
		http.NotFound(w, req)
		return nil
	}

	refs, err := sess.AdvertisedReferences()
	if err != nil {
		return err
	}

	refs.Capabilities.Add(capability.Sideband)
	refs.Capabilities.Add(capability.Sideband64k)

	w.Header().Add("Content-Type", fmt.Sprintf("application/x-%s-advertisement", svc))
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	enc := pktline.NewEncoder(w)
	enc.EncodeString(fmt.Sprintf("# service=%s\n", svc))
	enc.Flush()

	refs.Encode(w)
	return nil
}
