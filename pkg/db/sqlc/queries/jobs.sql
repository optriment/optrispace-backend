-- name: JobsList :many
select
    j.id,
    j.title,
    j.description,
    j.budget,
    j.duration,
    j.created_at,
    j.created_by,
    j.updated_at,
    p.address
    from jobs j
    left join persons p on j.created_by = p.id
    order by created_at;

-- name: JobGet :one
select
    j.id,
    j.title,
    j.description,
    j.budget,
    j.duration,
    j.created_at,
    j.created_by,
    j.updated_at,
    p.address,
    (select count(*) from applications a where a.job_id = j.id) as application_count
    from jobs j
    left join persons p on j.created_by = p.id
	where j.id = @id::varchar
    ;

-- name: JobAdd :one
insert into jobs (
    id, title, description, budget, duration, created_by
) values (
    $1, $2, $3, $4, $5, $6
) returning *;

-- name: JobsPurge :exec
-- Handle with care!
DELETE FROM jobs;