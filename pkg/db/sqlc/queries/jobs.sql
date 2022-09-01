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
    where j.blocked_at is null
    order by created_at desc;

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
    , COALESCE(p.display_name, p.login) AS customer_display_name
    from jobs j
    join persons p on p.id = j.created_by
    where j.id = @id::varchar and j.blocked_at is null;

-- name: JobAdd :one
insert into jobs (
    id, title, description, budget, duration, created_by
) values (
    $1, $2, $3, $4, $5, $6
) returning *;

-- name: JobPatch :one
update jobs
set
    title = case when @title_change::boolean
        then @title::varchar else title end,

    description = case when @description_change::boolean
        then @description::varchar else description end,

    budget = case when @budget_change::boolean
        then @budget::decimal else budget end,

    duration = case when @duration_change::boolean
        then @duration::int else duration end
where
    id = @id::varchar and @actor::varchar = created_by
returning *;

-- name: JobBlock :exec
update jobs
set
    blocked_at = now()
where
    id = @id::varchar
;

-- name: JobsPurge :exec
-- Handle with care!
DELETE FROM jobs;
