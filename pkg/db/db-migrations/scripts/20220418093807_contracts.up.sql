CREATE TABLE contracts (
    id varchar primary key not null,
    created_at timestamp not null default now(),
    updated_at  timestamp not null default now(),
    application_id varchar null references applications(id),
    performer_id varchar not null references persons(id),
    customer_id varchar not null references persons(id),
    title varchar not null,
    description text not null,
    price decimal not null,
    duration int null
);

comment on table contracts is 'Job offer table';

comment on column contracts.id is 'PK';
comment on column contracts.created_at is 'Creation timestamp';
comment on column contracts.updated_at is 'When the contract was updated last time';
comment on column contracts.application_id is 'Application was created before the contract';
comment on column contracts.performer_id is 'Person who performing the job';
comment on column contracts.customer_id is 'Customer for the job. Who paying.';
comment on column contracts.title is 'Contract title. Like "web site creation". Can be copied from the appropriate job.';
comment on column contracts.description is 'Details about the contract. May be long, long text. Also can be copied from the appropriate job.';
comment on column contracts.price is 'The crontract price';
comment on column contracts.duration is 'The contract duration';
