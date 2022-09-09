-- name: ChatGet :one
select
    c.*
from chats c
where c.id = @id::varchar
;

-- name: ChatGetByTopic :one
select
    c.*
from chats c
where c.topic = @topic::varchar
;

-- name: ChatAdd :one
insert into chats (
    id, topic
) values (
    @id, @topic
) returning *;

-- name: ChatParticipantAdd :one
insert into chats_participants (
    chat_id, person_id
) values (
    @chat_id, @person_id
) returning *;

-- name: ChatParticipantGet :one
select
    cp.*
from chats_participants cp
where
    cp.chat_id = @chat_id::varchar
    and cp.person_id = @person_id::varchar
;

-- name: ChatsPurge :exec
-- Handle with care!
DELETE FROM chats;