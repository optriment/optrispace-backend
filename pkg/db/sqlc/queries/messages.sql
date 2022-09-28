-- name: MessagesListByChat :many
select 
     m.*
    ,p.display_name
from messages m 
    join persons p on m.created_by = p.id
where m.chat_id = @chat_id::varchar
order by m.created_at asc;

-- name: MessageAdd :one
insert into messages (
    id, chat_id, created_by, text
) values (
    @id, @chat_id, @created_by, @text
) returning *;

-- name: MessagesPurge :exec
-- Handle with care!
DELETE FROM messages;