CREATE TABLE contracts (
    id varchar primary key not null,

    customer_id varchar not null references persons(id),
    performer_id varchar not null references persons(id),
    application_id varchar not null references applications(id),

    title varchar not null,
    description text not null,
    price decimal not null,
    duration int null,
    status varchar not null default 'created',

    created_by varchar not null references persons(id),
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),

    CONSTRAINT check_positive_price CHECK (price > 0.0)
);

create unique index contracts_customer_id_performer_id_application_id on public.contracts (customer_id, performer_id, application_id);

comment on table contracts is 'Contracts table';

comment on column contracts.id is 'PK';
comment on column contracts.customer_id is 'Customer for the job. Who paying.';
comment on column contracts.performer_id is 'Person who performing the job';
comment on column contracts.application_id is 'Application was created before the contract';
comment on column contracts.title is 'Contract title. Like "web site creation". Can be copied from the appropriate job.';
comment on column contracts.description is 'Details about the contract. May be long, long text. Also can be copied from the appropriate job.';
comment on column contracts.price is 'The crontract price';
comment on column contracts.duration is 'The contract duration';
comment on column contracts.status is 'Current status of the contract';
comment on column contracts.created_at is 'Creation timestamp';
comment on column contracts.updated_at is 'When the contract was updated last time';
