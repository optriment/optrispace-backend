-- name: ApplicationAdd :one
insert into applications (
    id, "comment", price, job_id, applicant_id
) values (
    $1, $2, $3, $4, $5
)
returning *;

-- name: ApplicationGet :one
select a.*
	, j.title AS job_title
	, j.budget AS job_budget
	, j.description AS job_description
	, c.id AS contract_id
	, c.status AS contract_status
	, (CASE WHEN p.display_name = '' THEN p.login ELSE p.display_name END)::varchar AS applicant_display_name
	, p.ethereum_address AS applicant_ethereum_address
	from applications a
	join jobs j on j.id = a.job_id
	join persons p on p.id = a.applicant_id
	left join contracts c on c.application_id  = a.id and c.performer_id = a.applicant_id
	where a.id = @id::varchar;

-- name: ApplicationsGetByJob :many
select a.*
	, c.id AS contract_id
	, c.status AS contract_status
	, (CASE WHEN p.display_name = '' THEN p.login ELSE p.display_name END)::varchar AS applicant_display_name
	, p.ethereum_address AS applicant_ethereum_address
	from applications a
	join persons p on p.id = a.applicant_id
	left join contracts c on c.application_id  = a.id and c.performer_id = a.applicant_id
	where a.job_id = @job_id::varchar
	order by a.created_at desc;

-- name: ApplicationFindByJobAndApplicant :one
select a.*
	, j.title AS job_title
	, j.budget AS job_budget
	, j.description AS job_description
	, c.id AS contract_id
	, c.status AS contract_status
	, (CASE WHEN p.display_name = '' THEN p.login ELSE p.display_name END)::varchar AS applicant_display_name
	, p.ethereum_address AS applicant_ethereum_address
	from applications a
	join persons p on p.id = a.applicant_id
	join jobs j on j.id = a.job_id
	left join contracts c on c.application_id  = a.id and c.performer_id = a.applicant_id
	where a.job_id = @job_id::varchar and a.applicant_id = @applicant_id::varchar
	limit 1;

-- name: ApplicationsGetByApplicant :many
select a.*
	, j.title AS job_title
	, j.budget AS job_budget
	, j.description AS job_description
	, c.id as contract_id
	, c.status as contract_status
	, c.price as contract_price
	, (CASE WHEN p.display_name = '' THEN p.login ELSE p.display_name END)::varchar AS applicant_display_name
	, p.ethereum_address AS applicant_ethereum_address
	from applications a
	join persons p on p.id = a.applicant_id
	join jobs j on a.job_id = j.id
	left join contracts c on a.id = c.application_id
	where a.applicant_id = @applicant_id::varchar
	order by a.created_at desc;

-- name: ApplicationsPurge :exec
-- Handle with care!
DELETE FROM applications;
