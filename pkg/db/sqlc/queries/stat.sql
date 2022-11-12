-- name: StatRegistrationsByDate :many
select date_trunc('day', p.created_at)::date "day", count(*) registrations
from persons p
group by day
order by day;

-- name: StatsGetOpenedJobsCount :one
select count(id) AS count
from jobs
where suspended_at is null and blocked_at is null;

-- name: StatsGetContractsCount :one
select count(id) AS count
from contracts;

-- name: StatsGetContractsVolume :one
select coalesce(sum(price), 0)::decimal AS volume
from contracts
where status IN ('approved', 'completed');
