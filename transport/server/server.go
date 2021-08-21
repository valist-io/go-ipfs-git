package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

var (
	ErrUpdateReference = errors.New("failed to update ref")
)

type session struct {
	storer storer.Storer
}

func NewReceivePackSession(storer storer.Storer) *receivePackSession {
	return &receivePackSession{
		session: session{storer},
	}
}

func NewUploadPackSession(storer storer.Storer) *uploadPackSession {
	return &uploadPackSession{
		session: session{storer},
	}
}

func (s *session) Close() error {
	return nil
}

func (s *session) AdvertisedReferencesContext(ctx context.Context) (*packp.AdvRefs, error) {
	return s.AdvertisedReferences()
}

func (s *session) AdvertisedReferences() (*packp.AdvRefs, error) {
	adv := packp.NewAdvRefs()
	if err := s.setReferences(adv); err != nil {
		return nil, err
	}

	if err := s.setHEAD(adv); err != nil {
		return nil, err
	}

	fmt.Println(adv.Head)
	return adv, nil
}

func (s *session) setReferences(adv *packp.AdvRefs) error {
	iter, err := s.storer.IterReferences()
	if err != nil {
		return err
	}

	return iter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() != plumbing.HashReference {
			return nil
		}

		adv.References[ref.Name().String()] = ref.Hash()
		return nil
	})
}

func (s *session) setHEAD(adv *packp.AdvRefs) error {
	ref, err := s.storer.Reference(plumbing.HEAD)
	if err == plumbing.ErrReferenceNotFound {
		return nil
	}

	if err != nil {
		return err
	}

	if ref.Type() == plumbing.SymbolicReference {
		if err := adv.AddReference(ref); err != nil {
			return nil
		}

		ref, err = storer.ResolveReference(s.storer, ref.Target())
		if err == plumbing.ErrReferenceNotFound {
			return nil
		}

		if err != nil {
			return err
		}
	}

	if ref.Type() != plumbing.HashReference {
		return plumbing.ErrInvalidType
	}

	h := ref.Hash()
	adv.Head = &h

	return nil
}
