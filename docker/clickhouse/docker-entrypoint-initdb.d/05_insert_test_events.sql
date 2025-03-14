insert into default.demo_events (dt, user_id, event, amount, screen)
select 
	now() - number % 10000000 as dt,
	cityHash64(number) % 1000000 as user_id,
	multiIf(
		cityHash64(number+2000) % 100 < 40, 'view',
		cityHash64(number+3000) % 1000 < 300, 'click',
		cityHash64(number+4000) % 5000 < 300, 'start_task',
	'other'
	) as event,
	0 as amount,
	multiIf(
		event IN ('start_task'), 'task',
		event IN ( 'view', 'click') AND cityHash64(number+2000) % 100 < 20, 'payment',
		event IN ( 'view', 'click') AND cityHash64(number+4000) % 500 < 70, 'task',
		'other'
	) as screen
from numbers(40000000);