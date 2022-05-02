-- name: ApplicationAdd :one
insert into applications (
    id, "comment", price, job_id, applicant_id
) values (
    $1, $2, $3, $4, $5
)
returning *;

-- on conflict
-- do nothing

-- name: ApplicationGet :one
select a.*, c.id as contract_id from applications a
	left join contracts c on a.id = c.application_id
	where a.id = @id::varchar;

-- name: ApplicationsListBy :many
select a.*, c.id as contract_id, c.status as contract_status, c.price as contract_price,
	j.title as job_title, j.description as job_description, j.budget as job_budget,
  COALESCE(p.display_name, p.login) AS applicant_display_name
	from applications a
	join jobs j on a.job_id = j.id
  join persons p on p.id = a.applicant_id
	left join contracts c on a.id = c.application_id
	where
	(@job_id::varchar = '' or a.job_id = @job_id::varchar)
	and (@actor_id::varchar = ''
		or a.applicant_id = @actor_id::varchar
		or j.created_by = @actor_id::varchar)
  order by a.created_at desc;

-- name: ApplicationsGetByJob :many
select * from applications
	where job_id = @job_id::varchar
  order by created_at desc;

-- name: ApplicationsGetByApplicant :many
select a.*, c.id as contract_id, c.status as contract_status, c.price as contract_price,
	j.title as job_title, j.description as job_description, j.budget as job_budget
	from applications a
	join jobs j on a.job_id = j.id
	left join contracts c on a.id = c.application_id
	where a.applicant_id = @applicant_id::varchar
  order by a.created_at desc;

-- name: ApplicationsPurge :exec
-- Handle with care!
DELETE FROM applications;
