// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: tests.sql

package pgdao

import (
	"context"
	"database/sql"
	"time"
)

const testsPersonCreate = `-- name: TestsPersonCreate :one
insert into persons (
    id, realm, login, password_hash, display_name, email, access_token, ethereum_address, is_admin, created_at
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
returning id, realm, login, password_hash, display_name, created_at, email, ethereum_address, resources, access_token, is_admin
`

type TestsPersonCreateParams struct {
	ID              string
	Realm           string
	Login           string
	PasswordHash    string
	DisplayName     string
	Email           string
	AccessToken     sql.NullString
	EthereumAddress string
	IsAdmin         bool
	CreatedAt       time.Time
}

func (q *Queries) TestsPersonCreate(ctx context.Context, arg TestsPersonCreateParams) (Person, error) {
	row := q.db.QueryRowContext(ctx, testsPersonCreate,
		arg.ID,
		arg.Realm,
		arg.Login,
		arg.PasswordHash,
		arg.DisplayName,
		arg.Email,
		arg.AccessToken,
		arg.EthereumAddress,
		arg.IsAdmin,
		arg.CreatedAt,
	)
	var i Person
	err := row.Scan(
		&i.ID,
		&i.Realm,
		&i.Login,
		&i.PasswordHash,
		&i.DisplayName,
		&i.CreatedAt,
		&i.Email,
		&i.EthereumAddress,
		&i.Resources,
		&i.AccessToken,
		&i.IsAdmin,
	)
	return i, err
}
