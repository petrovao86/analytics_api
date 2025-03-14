CREATE TABLE IF NOT EXISTS default.demo_events_buff AS default.demo_events 
ENGINE = Buffer(default, demo_events, 1, 10, 30, 10000, 1000000, 10000000, 100000000);