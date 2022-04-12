-- name: JobsList :many
select
    j.id,
    j.creation_ts,
    j.title,
    j.description,
    j.customer_id,
    p.address
    from jobs j
    right join persons p on j.customer_id = p.id
    order by creation_ts;

-- name: JobGet :one
select
    j.id,
    j.creation_ts,
    j.title,
    j.description,
    j.customer_id,
    p.address

    from jobs j
    right join persons p on j.customer_id = p.id
	where j.id = @id::varchar;

-- name: JobAdd :one
insert into jobs (
    id, title, description, customer_id
) values (
    $1, $2, $3, $4
) returning *;