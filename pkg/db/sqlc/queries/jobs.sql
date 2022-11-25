-- name: JobsList :many
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
    ,(CASE WHEN p.display_name = '' THEN p.login ELSE p.display_name END)::varchar AS customer_display_name
    ,p.ethereum_address AS customer_ethereum_address
    from jobs j
    join persons p on p.id = j.created_by
    where j.blocked_at is null and j.suspended_at is null and j.visibility = 'public'
    order by j.updated_at desc;

-- name: JobGet :one
select
    j.id
    ,j.title
    ,j.description
    ,j.budget
    ,j.duration
    ,j.created_at
    ,j.created_by
    ,j.updated_at
    ,j.suspended_at
    ,(select count(*) from applications a where a.job_id = j.id) as application_count
    ,(CASE WHEN p.display_name = '' THEN p.login ELSE p.display_name END)::varchar AS customer_display_name
    ,p.ethereum_address AS customer_ethereum_address
    from jobs j
    join persons p on p.id = j.created_by
    where j.id = @id::varchar and j.blocked_at is null;

-- name: JobFind :one
-- It is used only for testing purposes.
select * from jobs where id = @id::varchar;

-- name: JobAdd :one
insert into jobs (
    id, title, description, budget, duration, created_by
) values (
    $1, $2, $3, $4, $5, $6
) returning *;

-- name: JobPatch :one
update jobs
set
    title = @title::varchar,
    description = @description::varchar,
    budget = @budget::decimal,
    duration = @duration::int,
    updated_at = now()
where
    id = @id::varchar and @actor::varchar = created_by
returning *;

-- name: JobBlock :exec
update jobs set blocked_at = now() where id = @id::varchar;

-- name: JobSuspend :exec
update jobs set suspended_at = now() where id = @id::varchar;

-- name: JobResume :exec
update jobs set suspended_at = null where id = @id::varchar;

-- name: JobsPurge :exec
-- Handle with care!
DELETE FROM jobs;

-- name: JobHide :exec
update jobs set visibility = 'hidden' where id = @id::varchar;