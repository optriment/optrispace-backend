-- name: ContractAdd :one
insert into contracts (
    id, customer_id, performer_id, application_id, title, description, price, duration, created_by, customer_address, performer_address, status, contract_address
) values (
    @id, @customer_id, @performer_id, @application_id, @title, @description, @price, @duration, @created_by, @customer_address, @performer_address, @status, @contract_address
)
returning *;

-- name: ContractGet :one
-- mostly in testing purposes
select c.* from contracts c
join applications a on a.id = c.application_id and a.applicant_id = c.performer_id
join jobs j on j.id = a.job_id
join persons customer on customer.id = c.customer_id
join persons performer on performer.id = c.performer_id
where c.id = @id::varchar;

-- name: ContractGetByIDAndPersonID :one
select c.* from contracts c
join applications a on a.id = c.application_id and a.applicant_id = c.performer_id
join jobs j on j.id = a.job_id
join persons customer on customer.id = c.customer_id
join persons performer on performer.id = c.performer_id
where c.id = @id::varchar and (c.customer_id = @person_id::varchar or c.performer_id = @person_id::varchar);

-- name: ContractPatch :one
update contracts
set
    status = case when @status_change::boolean
        then @status::varchar else status end,

    performer_address = case when @performer_address_change::boolean
        then @performer_address::varchar else performer_address end,

    contract_address = case when @contract_address_change::boolean
        then @contract_address::varchar else contract_address end,

    updated_at = now()
where
    id = @id::varchar
returning *;

-- name: ContractsGetByPerson :many
select
    c.*
    ,(CASE WHEN pc.display_name = '' THEN pc.login ELSE pc.display_name END)::varchar AS customer_name
    ,(CASE WHEN pp.display_name = '' THEN pp.login ELSE pp.display_name END)::varchar AS performer_name
from contracts c
join persons pc on c.customer_id = pc.id
join persons pp on c.performer_id = pp.id
where c.customer_id = @person_id::varchar or c.performer_id = @person_id::varchar
order by c.created_at desc;

-- name: ContractsPurge :exec
-- Handle with care!
DELETE FROM contracts;
