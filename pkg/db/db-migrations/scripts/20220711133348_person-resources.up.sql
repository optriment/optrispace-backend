alter table persons
add column resources jsonb not null default '{}';

comment on column persons.resources is 'Person''s resources list. They may be links to social networks, portfolio, messenger IDs etc';