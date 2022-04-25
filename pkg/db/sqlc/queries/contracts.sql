-- name: ContractAdd :one
insert into contracts (
    id, customer_id, performer_id, application_id, title, description, price, duration, created_by
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
returning *;

-- on conflict
-- do nothing

-- name: ContractGetByIDAndPersonID :one
select c.* from contracts c
join applications a on a.id = c.application_id and a.applicant_id = c.performer_id
join jobs j on j.id = a.job_id
join persons customer on customer.id = c.customer_id
join persons performer on performer.id = c.performer_id
where c.id = @id::varchar and (c.customer_id = @person_id::varchar or c.performer_id = @person_id::varchar);

-- name: ContractsGetByPerson :many
select * from contracts
where customer_id = @person_id::varchar or performer_id = @person_id::varchar;

-- name: ContractsPurge :exec
-- Handle with care!
DELETE FROM contracts;
