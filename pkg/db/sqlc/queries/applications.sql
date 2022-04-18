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
select a.*, c.id as contract_id from applications a
	left join contracts c on a.id = c.application_id
	where 
	(@job_id::varchar = '' or a.job_id = @job_id::varchar) and 
	(@applicant_id::varchar = '' or a.applicant_id = @applicant_id::varchar);

-- name: ApplicationsGetByJob :many
select * from applications
	where job_id = @job_id::varchar;

-- name: ApplicationsGetByApplicant :many
select * from applications
	where applicant_id = @applicant_id::varchar;

-- name: ApplicationsPurge :exec
-- Handle with care!
DELETE FROM applications;