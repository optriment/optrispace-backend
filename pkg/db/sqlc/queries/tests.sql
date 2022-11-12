-- name: TestsPersonCreate :one
insert into persons (
    id, realm, login, password_hash, display_name, email, access_token, ethereum_address, is_admin, created_at
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
returning *;
