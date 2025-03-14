CREATE TABLE IF NOT EXISTS default.ch_migrations (
    name String, 
    group_id Int64, 
    migrated_at DateTime, 
    sign Int8
) Engine = CollapsingMergeTree(sign) 
ORDER BY (name);
