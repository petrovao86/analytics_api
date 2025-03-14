insert into default.demo_events (dt, user_id, event, amount, screen)
select 
	dt + cityHash64(dt) % 5*60 as dt,
	user_id,
	'pay' as event,
	100 as amount,
	screen
from default.demo_events as e
where  e.event = 'view' and screen = 'payment' and cityHash64(dt, user_id) % 1000 < 60;