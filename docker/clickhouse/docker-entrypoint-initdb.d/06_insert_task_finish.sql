insert into default.demo_events (dt, user_id, event, amount, screen)
select 
	dt + cityHash64(dt) % 45000 as dt,
	user_id,
	'finish_task' as event,
	0 as amount,
	'task' as screen
from default.demo_events as e
where  e.event = 'start_task' and cityHash64(dt, user_id) % 1000 < 250;