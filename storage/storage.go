package storage

import (
	"context"

	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/valist-io/go-ipfs-git/storage/unixfs"
)

// Storage is an implementation of git.Storer
type Storage struct {
	fs *unixfs.Unixfs
	ConfigStorage
	ObjectStorage
	ShallowStorage
	IndexStorage
	ReferenceStorage
	ModuleStorage
}

func newStorage(fs *unixfs.Unixfs) *Storage {
	return &Storage{
		fs:               fs,
		ConfigStorage:    ConfigStorage{},
		ShallowStorage:   ShallowStorage{},
		ReferenceStorage: NewReferenceStorage(fs),
		ObjectStorage:    NewObjectStorage(fs),
		ModuleStorage:    NewModuleStorage(fs),
	}
}

func NewStorage(ctx context.Context, ds ipld.DAGService) (*Storage, error) {
	fs, err := unixfs.New(ctx, ds)
	if err != nil {
		return nil, err
	}

	return newStorage(fs), nil
}

func LoadStorage(ctx context.Context, ds ipld.DAGService, id cid.Cid) (*Storage, error) {
	fs, err := unixfs.Load(ctx, ds, id)
	if err != nil {
		return nil, err
	}

	return newStorage(fs), nil
}

func (s *Storage) Node() (ipld.Node, error) {
	return s.fs.Node()
}
