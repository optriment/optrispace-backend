-- name: PersonAdd :one
insert into persons (
    id, address
) values (
    $1, $2
) 
on conflict (address)
do 
   update set address = $2
returning *;

-- name: PersonGet :one
select * from persons
	where id = @id::varchar;

-- name: PersonGetByAddress :one
select * from persons
	where address = @address::varchar;

-- name: PersonsList :many
select * from persons
	where id = @id::varchar;
