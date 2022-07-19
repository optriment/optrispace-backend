-- name: StatRegistrationsByDate :many
select date_trunc('day', p.created_at)::date "day", count(*) registrations
from persons p
group by day
order by day;
