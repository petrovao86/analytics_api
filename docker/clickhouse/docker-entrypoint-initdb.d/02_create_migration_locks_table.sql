CREATE TABLE IF NOT EXISTS default.ch_migration_locks (
    a Int8
) Engine = MergeTree() 
ORDER BY tuple();
