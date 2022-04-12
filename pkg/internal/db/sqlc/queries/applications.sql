-- name: ApplicationAdd :one
insert into applications (
    id, job_id, applicant_id
) values (
    $1, $2, $3
)
returning *;

-- name: ApplicationGet :one
select * from applications
	where id = @id::varchar;

-- name: ApplicationsGetByJob :many
select * from applications
	where job_id = @job_id::varchar;

-- name: ApplicationsGetByApplicant :many
select * from applications
	where applicant_id = @applicant_id::varchar;
