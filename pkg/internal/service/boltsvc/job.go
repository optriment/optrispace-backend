package boltsvc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
	"optrispace.com/work/pkg/model"
)

type (
	jobService struct {
		db *bolt.DB
	}
)

const jobBucketName = "job"

// NewJob creates new Job service with bolt DB persistent implementation
// path has to point to real file path for store data
func NewJob(ctx context.Context, path string) (*jobService, error) {
	db, err := openBolt(ctx, path)
	if err != nil {
		return nil, err
	}

	return &jobService{
		db: db,
	}, nil
}

// New saves job into bolt DB.
// It sets new ID if unspecified.
func (s *jobService) Save(ctx context.Context, job *model.Job) error {
	if job.ID == "" {
		job.ID = newID()
	}

	job.CreationTime = time.Now()

	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(jobBucketName))
		if err != nil {
			return err
		}

		bb, err := json.Marshal(job)
		if err != nil {
			return err
		}

		return b.Put([]byte(job.ID), bb)
	})
}

func (s *jobService) Read(ctx context.Context, id string) (*model.Job, error) {
	job := new(model.Job)
	return job, s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(jobBucketName))
		if b == nil {
			return model.ErrEntityNotFound
		}

		bb := b.Get([]byte(id))
		if bb == nil {
			return model.ErrEntityNotFound
		}

		return json.Unmarshal(bb, job)
	})
}

func (s *jobService) ReadAll(ctx context.Context) ([]*model.Job, error) {
	jobs := make([]*model.Job, 0, 0)

	e := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(jobBucketName))
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			job := new(model.Job)

			err := json.Unmarshal(v, job)
			if err != nil {
				log.Warn().Err(err).Str("key", string(k)).Msg("Unable to unmarshal JSON")
				return nil // proceed!
			}

			if job.ID != string(k) {
				log.Warn().Str("key", string(k)).Msg("Object key does not conform ID within the object.")
				job.ID = string(k)
			}

			jobs = append(jobs, job)

			return nil
		})
	})

	return jobs, e
}
