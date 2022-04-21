-- name: PersonAdd :one
insert into persons (
    id, realm, login, password_hash, display_name, email
) values (
    $1, $2, $3, $4, $5, $6
) 
returning *;

-- name: PersonGet :one
select * from persons
	where id = @id::varchar;

-- name: PersonGetByLogin :one
select * from persons p
	where p.login = @login::varchar and p.realm = @realm::varchar;

-- name: PersonsList :many
select * from persons;

-- name: PersonsPurge :exec
-- Handle with care!
DELETE FROM persons;