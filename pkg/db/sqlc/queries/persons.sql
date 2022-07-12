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

-- name: PersonSetPassword :exec
update persons
set
    password_hash = @new_password_hash::varchar
where
    id = @id::varchar
;

-- name: PersonSetResources :exec
update persons
set
    resources = @resources::json
where
    id = @id::varchar
;

-- name: PersonPatch :one
update persons
set
    ethereum_address = case when @ethereum_address_change::boolean
        then @ethereum_address::varchar else ethereum_address end

    , display_name = case when @display_name_change::boolean
        then @display_name::varchar else display_name end

where
    id = @id::varchar
returning *;


-- name: PersonsPurge :exec
-- Handle with care!
DELETE FROM persons;