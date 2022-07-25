// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: stat.sql

package pgdao

import (
	"context"
	"time"
)

const statRegistrationsByDate = `-- name: StatRegistrationsByDate :many
select date_trunc('day', p.created_at)::date "day", count(*) registrations
from persons p
group by day
order by day
`

type StatRegistrationsByDateRow struct {
	Day           time.Time
	Registrations int64
}

func (q *Queries) StatRegistrationsByDate(ctx context.Context) ([]StatRegistrationsByDateRow, error) {
	rows, err := q.db.QueryContext(ctx, statRegistrationsByDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StatRegistrationsByDateRow
	for rows.Next() {
		var i StatRegistrationsByDateRow
		if err := rows.Scan(&i.Day, &i.Registrations); err != nil {
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