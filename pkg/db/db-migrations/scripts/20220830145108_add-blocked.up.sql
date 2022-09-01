alter table jobs
add column blocked_at timestamp null default null;

comment on column jobs.blocked_at is 'Job is blocked if this field is not null';
