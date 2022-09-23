create table chats (
    id varchar primary key not null
    , topic varchar not null check (starts_with(topic, 'urn:application:') or starts_with(topic, 'urn:contract:')) unique
    , created_at timestamp not null default now()
);

comment on table chats is 'Chats where users have conversations';

comment on column chats.id is 'PK';
comment on column chats.topic is 'Topic what is talk about. In form of URI in form: urn:<type>:id. Where is type is: application, contract etc. ID is id of appropriate entity.';
comment on column chats.created_at is 'Creation timestamp';

create table messages (
    id varchar primary key not null
    , chat_id varchar not null references chats(id)
    , created_at timestamp not null default now()
    , created_by varchar not null references persons(id)
    , "text" text not null
);

comment on table messages is 'Messages were sent in chats by users';

comment on column messages.id is 'PK';
comment on column messages.chat_id is 'Chat where message was sent';
comment on column messages.created_at is 'Creation timestamp';
comment on column messages.created_by is 'User who sent message';
comment on column messages.text is 'It is a message text, in fact';

comment on column contracts.price is 'The contract price';

create table chats_participants (
      chat_id varchar not null references chats(id) on delete cascade
    , person_id varchar not null references persons(id) on delete cascade
);

create unique index chats_participants_idx on chats_participants (chat_id,person_id);

comment on table chats_participants is 'Participants in chats';

comment on column chats_participants.chat_id is 'Chat where user joined';
comment on column chats_participants.person_id is 'User';

comment on column contracts.created_by is 'Customer';
