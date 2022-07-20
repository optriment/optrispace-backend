alter table persons
add column access_token varchar null unique;

comment on column persons.access_token is 'Person''s personal access token for Bearer authentication schema';