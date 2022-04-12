package pgsvc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ryboe/q"
	"optrispace.com/work/pkg/internal/db/pgdao"
	"optrispace.com/work/pkg/model"
)

type (
	jobSvc struct {
		db *sql.DB
	}
)

func NewJob(db *sql.DB) *jobSvc {
	return &jobSvc{db: db}
}

// Add implements service.Job interface
func (s *jobSvc) Add(ctx context.Context, job *model.Job) (*model.Job, error) {
	var result *model.Job
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		id := job.ID
		if id == "" {
			id = newID()
		}

		if job.Customer.ID == "" {

			if job.Customer.Address == "" {
				return fmt.Errorf("address: %w", model.ErrRequiredFieldNotFilled)
			}

			p, err := queries.PersonGetByAddress(ctx, job.Customer.Address)
			if err != nil {
				return fmt.Errorf("unable to PersonGetByAddress: %w", err)
			}

			job.Customer = &model.Person{
				ID:      p.ID,
				Address: p.Address,
			}

		}

		j, err := queries.JobAdd(ctx,
			pgdao.JobAddParams{
				ID:          id,
				Title:       job.Title,
				Description: job.Description,
				CustomerID:  job.Customer.ID,
			})
		if err != nil {
			return fmt.Errorf("unable to JobAdd job: %w", err)
		}

		result = &model.Job{
			ID:           j.ID,
			CreationTime: j.CreationTs,
			Title:        j.Title,
			Description:  j.Description,
			Customer:     job.Customer,
		}
		q.Q(result)
		return nil
	})
}

// Get implements service.Job interface
func (s *jobSvc) Get(ctx context.Context, id string) (*model.Job, error) {
	var result *model.Job
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		j, err := queries.JobGet(ctx, id)
		if err != nil {
			return fmt.Errorf("unable to JobGet jobs with id='%s': %w", id, err)
		}

		result = &model.Job{
			ID:           j.ID.String,
			CreationTime: j.CreationTs.Time,
			Title:        j.Title.String,
			Description:  j.Description.String,
			Customer: &model.Person{
				ID:      j.CustomerID.String,
				Address: j.Address,
			},
		}
		return nil
	})
}

// List implements service.Job interface
func (s *jobSvc) List(ctx context.Context) ([]*model.Job, error) {
	var result []*model.Job = make([]*model.Job, 0)
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		jj, err := queries.JobsList(ctx)
		if err != nil {
			return fmt.Errorf("unable to JobReadAll job: %w", err)
		}
		for _, j := range jj {
			result = append(result, &model.Job{
				ID:           j.ID.String,
				CreationTime: j.CreationTs.Time,
				Title:        j.Title.String,
				Description:  j.Description.String,
				Customer: &model.Person{
					ID:      j.CustomerID.String,
					Address: j.Address,
				},
			})
		}
		return nil
	})
}
