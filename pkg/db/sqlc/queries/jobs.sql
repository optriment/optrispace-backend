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
    from jobs j
    order by created_at;

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
    ,(select count(*) from applications a where a.job_id = j.id) as application_count
    from jobs j
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
