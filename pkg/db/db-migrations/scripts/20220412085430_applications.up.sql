CREATE TABLE applications (
    id varchar primary key not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    "comment" text not null,
    price decimal not null,
    job_id varchar not null references jobs(id),
    applicant_id varchar not null references persons(id)
);

create unique index applications_job_id_applicant_id on public.applications (job_id,applicant_id);

comment on table applications is 'Applications for job offers';

comment on column applications.id is 'PK';
comment on column applications.created_at is 'Application timestamp. When application was created.';
comment on column applications.updated_at is 'Application update timestamp. When application was updated last time.';
comment on column applications.comment is 'Applicant''s initial comment on the application';
comment on column applications.price is 'Proposed price';
comment on column applications.job_id is 'Job offer';
comment on column applications.applicant_id is 'Potential performer';
