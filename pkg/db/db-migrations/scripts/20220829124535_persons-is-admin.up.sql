alter table persons
add column is_admin boolean not null default false;

comment on column persons.is_admin is 'Does user have admin privileges?';