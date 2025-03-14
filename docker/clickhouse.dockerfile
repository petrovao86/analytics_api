FROM clickhouse/clickhouse-server:24.12-alpine

ADD clickhouse/config.xml /etc/clickhouse-server/config.xml
ADD clickhouse/users.xml /var/lib/clickhouse/user_files/users.xml
