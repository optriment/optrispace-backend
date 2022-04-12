CREATE TABLE persons (
    id varchar primary key not null,

    address varchar not null UNIQUE
);

comment on table persons is 'Person who can pay, get or earn money';

comment on column persons.id is 'PK';
comment on column persons.address is 'This person''s cryptocurrency address';
