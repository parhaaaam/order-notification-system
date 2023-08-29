-- name: GetAllDelaysInLastWeek :many
select v.slug, sum(extract(epoch from o.delivered_at - o.time_delivery)) as delay_amount
from vendor v
         join "order" o on v.id = o.vendor_id
         join delay_report d on d.order_id = o.id
where o.created_at < now() - interval '1 week'
group by v.slug
order by 2 desc;

-- name: GetTripStatusAndOrderTimeDeliveryByOrderId :one
select t.status, o.time_delivery
from "order" o
         join "trip" t on o.id = t.order_id
where o.id = $1;

-- name: AddDelayReports :one
insert into delay_report (description, order_id)
values ($1, $2)
returning id;

-- name: AssignOrderToAgent :exec
update delay_report
set agent_id = $1
where id = $2;

-- name: CheckDelayReportOrderIDIsClosed :one
select case when count(*) = 0 then true else false end as all_closed
from delay_report
where order_id = $1 and status != 'CLOSED';