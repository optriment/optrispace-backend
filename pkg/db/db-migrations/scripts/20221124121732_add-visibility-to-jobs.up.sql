alter table jobs
add column visibility varchar(20) not null default 'public';

create index jobs_visibility on jobs (visibility);

comment on column jobs.visibility is 'Job visibility. Like public and hidden.';