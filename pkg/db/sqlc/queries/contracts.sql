-- name: ContractAdd :one
insert into contracts (
    id, customer_id, performer_id, application_id, title, description, price, duration, created_by
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
returning *;

-- on conflict
-- do nothing

-- name: ContractsPurge :exec
-- Handle with care!
DELETE FROM contracts;
