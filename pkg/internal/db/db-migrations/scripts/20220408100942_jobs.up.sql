CREATE TABLE jobs (
    id varchar primary key not null,

    creation_ts timestamp not null default now(),

    title varchar not null,

    description text not null,

    customer_id varchar not null references persons(id)

);

comment on table jobs is 'Job offer table';

comment on column jobs.id is 'PK';
comment on column jobs.creation_ts is 'Creation timestamp';

comment on column jobs.title is 'Job title. Like "web site creation"';
comment on column jobs.description is 'Details for these job. May be long, long text';
comment on column jobs.customer_id is 'Who will pay this job';
