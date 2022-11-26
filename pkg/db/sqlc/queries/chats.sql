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

-- name: ChatsListByParticipant :many
-- Chat list order by created_at timestamp of the latest message that belongs chat. Order is inverse from the newest to the latest update.
select
      c.*
    , p.id as person_id
    , p.display_name as person_display_name
    , p.ethereum_address as person_ethereum_address
    , m1.created_at as last_message_at
from chats c
join chats_participants cp on c.id = cp.chat_id
join persons p on p.id = cp.person_id
join messages m1 on c.id = m1.chat_id
left outer join messages m2 on (c.id = m2.chat_id and m1.created_at < m2.created_at)
where c.id in (select chat_id from chats_participants where person_id = @participant_id::varchar)
and m2.id is null
order by m1.created_at desc, c.id
;

-- name: ChatGetDetailsByApplicationID :one
select
      a.id as application_id
    , j.id as job_id
    , j.title as job_title
    , c.id as contract_id
from applications a
join jobs j on a.job_id = j.id
left join contracts c on a.id = c.application_id
where a.id = @application_id::varchar
;


-- name: ChatsPurge :exec
-- Handle with care!
DELETE FROM chats;
