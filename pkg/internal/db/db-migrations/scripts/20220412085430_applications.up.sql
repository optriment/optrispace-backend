CREATE TABLE applications (
    id varchar primary key not null,

    creation_ts timestamp not null default now(),

    job_id varchar not null references jobs(id),

    applicant_id varchar not null references persons(id)
);

create unique index applications_job_id_applicant_id on public.applications (job_id,applicant_id);

comment on table applications is 'Applications for job offers';

comment on column applications.id is 'PK';
comment on column applications.creation_ts is 'Application timestamp. When application was created.';

comment on column applications.job_id is 'Job offer';
comment on column applications.applicant_id is 'Potential performer';
