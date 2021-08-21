package storage

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5/storage/test"
	"github.com/ipfs/go-merkledag/dagutils"
	check "gopkg.in/check.v1"
)

var _ = check.Suite(&StorageSuite{})

type StorageSuite struct {
	test.BaseStorageSuite
}

func (s *StorageSuite) SetUpTest(c *check.C) {
	ctx := context.Background()
	ds := dagutils.NewMemoryDagService()

	storer, err := NewStorage(ctx, ds)
	if err != nil {
		c.Fatalf("failed to create storage: %v", err)
	}

	s.BaseStorageSuite = test.NewBaseStorageSuite(storer)
}

func Test(t *testing.T) {
	check.TestingT(t)
}
