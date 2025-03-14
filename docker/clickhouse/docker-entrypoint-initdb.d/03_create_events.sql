CREATE TABLE IF NOT EXISTS default.demo_events (
    dt      DateTime,
    event   LowCardinality(String),
    user_id String,
    screen  String,
    elem    String,
    amount  Int64
)
ENGINE = MergeTree() 
PARTITION BY toYYYYMM(dt)
ORDER BY (dt, event);