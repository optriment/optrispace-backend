CREATE TABLE persons (
    id varchar primary key not null,
    realm varchar not null,
    login varchar not null,
    password_hash varchar not null,
    display_name varchar not null,
    created_at timestamp not null default now(),
    email varchar not null
);

create unique index persons_realm_login on persons (realm,login);

comment on table persons is 'Person who can pay, get or earn money';

comment on column persons.id is 'PK';
comment on column persons.realm is 'Authentication realm (inhouse by default)';
comment on column persons.login is 'Login for authentication (must be unique for the separate authentication realm)';
comment on column persons.password_hash is 'Salty password hash';
comment on column persons.display_name is 'User name for displaying';
comment on column persons.created_at is 'Creation time';
comment on column persons.email is 'User Email';
