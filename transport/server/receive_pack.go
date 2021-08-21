package server

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
)

type receivePackSession struct {
	session
	err error
}

func (s *receivePackSession) ReceivePack(ctx context.Context, req *packp.ReferenceUpdateRequest) (*packp.ReportStatus, error) {
	rs := packp.NewReportStatus()
	rs.UnpackStatus = "ok"

	if err := packfile.UpdateObjectStorage(s.storer, req.Packfile); err != nil {
		rs.UnpackStatus = err.Error()
		return rs, err
	}

	s.updateReferences(rs, req.Commands)
	if req.Capabilities.Supports(capability.ReportStatus) {
		return rs, s.err
	}

	return nil, s.err
}

func (s *receivePackSession) updateReferences(rs *packp.ReportStatus, commands []*packp.Command) {
	for _, cmd := range commands {
		exists, err := s.referenceExists(cmd.Name)
		if err != nil {
			s.setStatus(rs, cmd.Name, err)
			continue
		}

		switch cmd.Action() {
		case packp.Create:
			if exists {
				s.setStatus(rs, cmd.Name, ErrUpdateReference)
				continue
			}

			ref := plumbing.NewHashReference(cmd.Name, cmd.New)
			err := s.storer.SetReference(ref)
			s.setStatus(rs, cmd.Name, err)
		case packp.Delete:
			if !exists {
				s.setStatus(rs, cmd.Name, ErrUpdateReference)
				continue
			}

			err := s.storer.RemoveReference(cmd.Name)
			s.setStatus(rs, cmd.Name, err)
		case packp.Update:
			if !exists {
				s.setStatus(rs, cmd.Name, ErrUpdateReference)
				continue
			}

			ref := plumbing.NewHashReference(cmd.Name, cmd.New)
			err := s.storer.SetReference(ref)
			s.setStatus(rs, cmd.Name, err)
		}
	}
}

func (s *receivePackSession) setStatus(rs *packp.ReportStatus, ref plumbing.ReferenceName, err error) {
	status := &packp.CommandStatus{
		ReferenceName: ref,
		Status:        "ok",
	}

	if err != nil {
		status.Status = err.Error()
	}

	if s.err != nil {
		s.err = err
	}

	rs.CommandStatuses = append(rs.CommandStatuses, status)
}

func (s *receivePackSession) referenceExists(n plumbing.ReferenceName) (bool, error) {
	_, err := s.storer.Reference(n)
	if err == plumbing.ErrReferenceNotFound {
		return false, nil
	}

	return err == nil, err
}
