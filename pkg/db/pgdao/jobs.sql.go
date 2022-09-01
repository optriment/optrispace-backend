// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: jobs.sql

package pgdao

import (
	"context"
	"database/sql"
	"time"
)

const jobAdd = `-- name: JobAdd :one
insert into jobs (
    id, title, description, budget, duration, created_by
) values (
    $1, $2, $3, $4, $5, $6
) returning id, title, description, budget, duration, created_at, updated_at, created_by, blocked_at
`

type JobAddParams struct {
	ID          string
	Title       string
	Description string
	Budget      sql.NullString
	Duration    sql.NullInt32
	CreatedBy   string
}

func (q *Queries) JobAdd(ctx context.Context, arg JobAddParams) (Job, error) {
	row := q.db.QueryRowContext(ctx, jobAdd,
		arg.ID,
		arg.Title,
		arg.Description,
		arg.Budget,
		arg.Duration,
		arg.CreatedBy,
	)
	var i Job
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.Budget,
		&i.Duration,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.BlockedAt,
	)
	return i, err
}

const jobBlock = `-- name: JobBlock :exec
update jobs
set
    blocked_at = now()
where
    id = $1::varchar
`

func (q *Queries) JobBlock(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, jobBlock, id)
	return err
}

const jobGet = `-- name: JobGet :one
select
    j.id
    ,j.title
    ,j.description
    ,j.budget
    ,j.duration
    ,j.created_at
    ,j.created_by
    ,j.updated_at
    ,(select count(*) from applications a where a.job_id = j.id) as application_count
    , COALESCE(p.display_name, p.login) AS customer_display_name
    from jobs j
    join persons p on p.id = j.created_by
    where j.id = $1::varchar and j.blocked_at is null
`

type JobGetRow struct {
	ID                  string
	Title               string
	Description         string
	Budget              sql.NullString
	Duration            sql.NullInt32
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	ApplicationCount    int64
	CustomerDisplayName string
}

func (q *Queries) JobGet(ctx context.Context, id string) (JobGetRow, error) {
	row := q.db.QueryRowContext(ctx, jobGet, id)
	var i JobGetRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.Budget,
		&i.Duration,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.ApplicationCount,
		&i.CustomerDisplayName,
	)
	return i, err
}

const jobPatch = `-- name: JobPatch :one
update jobs
set
    title = case when $1::boolean
        then $2::varchar else title end,

    description = case when $3::boolean
        then $4::varchar else description end,

    budget = case when $5::boolean
        then $6::decimal else budget end,

    duration = case when $7::boolean
        then $8::int else duration end
where
    id = $9::varchar and $10::varchar = created_by
returning id, title, description, budget, duration, created_at, updated_at, created_by, blocked_at
`

type JobPatchParams struct {
	TitleChange       bool
	Title             string
	DescriptionChange bool
	Description       string
	BudgetChange      bool
	Budget            string
	DurationChange    bool
	Duration          int32
	ID                string
	Actor             string
}

func (q *Queries) JobPatch(ctx context.Context, arg JobPatchParams) (Job, error) {
	row := q.db.QueryRowContext(ctx, jobPatch,
		arg.TitleChange,
		arg.Title,
		arg.DescriptionChange,
		arg.Description,
		arg.BudgetChange,
		arg.Budget,
		arg.DurationChange,
		arg.Duration,
		arg.ID,
		arg.Actor,
	)
	var i Job
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.Budget,
		&i.Duration,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.BlockedAt,
	)
	return i, err
}

const jobsList = `-- name: JobsList :many
select
     j.id
    ,j.title
    ,j.description
    ,j.budget
    ,j.duration
    ,j.created_at
    ,j.created_by
    ,j.updated_at
    ,(select count(*) from applications a where a.job_id = j.id) as application_count
    from jobs j
    where j.blocked_at is null
    order by created_at desc
`

type JobsListRow struct {
	ID               string
	Title            string
	Description      string
	Budget           sql.NullString
	Duration         sql.NullInt32
	CreatedAt        time.Time
	CreatedBy        string
	UpdatedAt        time.Time
	ApplicationCount int64
}

func (q *Queries) JobsList(ctx context.Context) ([]JobsListRow, error) {
	rows, err := q.db.QueryContext(ctx, jobsList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []JobsListRow
	for rows.Next() {
		var i JobsListRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.Budget,
			&i.Duration,
			&i.CreatedAt,
			&i.CreatedBy,
			&i.UpdatedAt,
			&i.ApplicationCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const jobsPurge = `-- name: JobsPurge :exec
DELETE FROM jobs
`

// Handle with care!
func (q *Queries) JobsPurge(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, jobsPurge)
	return err
}
