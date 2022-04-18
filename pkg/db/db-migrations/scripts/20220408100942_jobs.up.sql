CREATE TABLE jobs (
    id varchar primary key not null,
    title varchar not null,
    description text not null,
    budget decimal null,
    duration int null,
    created_at timestamp not null default now(),
    updated_at  timestamp not null default now(),
    created_by varchar not null references persons(id)
);

comment on table jobs is 'Job offer table';

comment on column jobs.id is 'PK';
comment on column jobs.title is 'Job title. Like "web site creation"';
comment on column jobs.description is 'Details about the job. May be long, long text';
comment on column jobs.budget is 'Estimated cost of the job if any';
comment on column jobs.duration is 'Estimated duration of the job in days if any';
comment on column jobs.created_at is 'Creation timestamp';
comment on column jobs.updated_at is 'When the job was updated last time';
comment on column jobs.created_by is 'Who created this job and should pay for it';
