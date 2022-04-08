package service

import (
	"context"

	"optrispace.com/work/pkg/internal/service/boltsvc"
	"optrispace.com/work/pkg/model"
)

type (
	Job interface {
		// Save saves the job into storage
		Save(ctx context.Context, job *model.Job) error
		// Read reads specified job from storage by specified id
		// It can return model.ErrNotFound
		Read(ctx context.Context, id string) (*model.Job, error)

		// Reads all items from storage and returns result to
		ReadAll(ctx context.Context) ([]*model.Job, error)
	}
)

func NewJobService(ctx context.Context, path string) (Job, error) {
	return boltsvc.NewJob(ctx, path)
}
