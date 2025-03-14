# Демо сервиса сбора событий
Демо архитектуры сервиса аналитики отвечающего за приём и обработку событий.
### Запуск

```shell
git clone https://github.com/petrovao86/analytics_api
cd analytics_api
docker compose -f ./docker/docker-compose.yaml -p analytics_api up --build -d
```

### Остановка
```shell
docker compose -f ./docker/docker-compose.yaml -p logic_like down --volumes
```


### Описание
- [Архитектура](#архитектура)
- [API](#api)
- [Clickhouse](#clickhouse)
- [Генератор нагрузки](#генератор-нагрузки)
- [Grafana](#grafana)

### Архитектура

![Архитектура](./docs/images/arch.svg)
#### Описание
С учётом заявленных нагрузок порядка 230 rps(пики скорее всего до 500) и требования эфективного использования ресурсов, выбрана следующая схема обработки потока событий:
1. События принимаются в [API](#api) 
2. [API](#api) парсит и валидирует события
3. [API](#api) вставляется события в [буферную таблицу](#буферная-таблица) Clickhouse
4. [Буферная таблица](#буферная-таблица) периодически скидывает данные в [основную таблицу](#основная-таблица)
5. Все необходимые витрины формируются материальными представлениями на стороне Кликхауса и/или etl-задачами запускаемыми в [celery](https://github.com/celery/celery), [airflow](https://github.com/apache/airflow), [prefect](https://github.com/PrefectHQ/prefect), [dagster](https://github.com/dagster-io/dagster), [machinery](https://github.com/RichardKnop/machinery) и т.п.


### API

Адрес API `http://localhost:18888/events/`

Структура запроса:
| Поле |Тип данных| Описание                                                               |
|------|----------|:-----------------------------------------------------------------------|
|  dt  | DateTime | Время события. Формат `2020-01-01T14:16:34Z`. **Обязательное**.        |
|event |  String  | Название события. **Обязательное**.                                    |
|userid|  String  | Идентификатор пользователя. **Обязательное**.                          |
|screen|  String  | Экран на котором произошло событие. Например: `payment`, `login` и т.д.|
| elem |  String  | Объект с которым произошло событие. Например: `save_button`, `main_screen`, `pay_button` и т.п.|
|amount|   Int    | Поле для указание сумм, если событие связано с оплатой или стоимостью чего-либо. При просмотре курса стоимостью в 100 р. указывается 100, при продлении прописки за 500 р. указывается 500|

___
Пример оправки события:
```shell
 curl -X POST "http://localhost:18888/events/" -F dt=2025-03-01T11:11:11Z -F event=test -F userid=1
```
Просмотр главного экрана:
```shell
curl -X POST "http://localhost:18888/events/" -H "Content-Type: application/json" --data '{"dt":"2020-01-01T14:16:34Z", "userid": "1", "event": "view", "screen": "main", "elem": "screen"}'
```
Просмотр курса:
```shell
curl -X POST "http://localhost:18888/events/" -H "Content-Type: application/json" --data '{"dt":"2020-01-01T14:16:34Z", "userid": "1", "event": "view", "screen": "course", "elem": "screen", "amount": 100}'
```
Оплата:
```shell
curl -X POST "http://localhost:18888/events/" -F dt=2020-01-01T14:16:34Z -F event=pay -F userid=1 -F screen=payment -F amount=100
```


### ClickHouse
Логин `user`, пароль `pass`

- web-интерфейс: http://127.0.0.1:18123/play?user=user&password=pass
- http api: `127.0.0.1:18123`
- tcp api: `127.0.0.1:19000`

#### Схема данных
##### [Основная таблица](./docker/clickhouse/docker-entrypoint-initdb.d/03_create_events.sqls)
```sql
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
```
Т.к. ClickHouse не любит точечные вставки, поток событий перед записью необходимо буферизировать. 

Возможные варианты:
1. Копить события прямо в API
2. Копить события в очереди на `Redis`, `Aerospike`, `Kafka`, `NATS`
3. Копить события [средствами ClickHouse](https://clickhouse.com/docs/engines/table-engines/special/buffer) в [буферной таблице](#буферная-таблица)

С учётом объёма событий, самый простой и эффективный вариант - писать в [буферную таблицу](#буферная-таблица), это позволит оставить API простым и не переусложнит архитектуру решения дополнительными сервисами.

##### [Буферная таблица](./docker/clickhouse/docker-entrypoint-initdb.d/04_create_events_buffer.sql)
```sql
CREATE TABLE IF NOT EXISTS default.demo_events_buff AS default.demo_events 
ENGINE = Buffer(default, demo_events, 1, 10, 30, 10000, 1000000, 10000000, 100000000);
```
#### Построение отчётов
При старте сервера таблица заполняется демонстрационными данными.

Можно попробовать различные запросы:
###### dau, wau, mau и т.п. в скользящем по дням окне
```sql
select
	date,
	uniqMerge(dauState) OVER (ORDER BY date RANGE BETWEEN CURRENT ROW AND CURRENT ROW) as dau,
	uniqMerge(dauState) OVER (ORDER BY date RANGE BETWEEN 7 PRECEDING AND CURRENT ROW) as wau,
	uniqMerge(dauState) OVER (ORDER BY date RANGE BETWEEN 30 PRECEDING AND CURRENT ROW) as mau
FROM (
	select
		toDate(dt) as date,
		uniqState(user_id) as dauState
	from default.demo_events
	where dt >= now() - toIntervalDay(6*30)
	group by date
)
```
###### средний процент решения задач
```sql
select
	date,
	avg(finished_tasks/started_tasks) as solve_ratio
from (
	select
		toDate(dt) as date,
		user_id,
		sum(event='finish_task') as finished_tasks,
		sum(event='start_task') as started_tasks
	from default.demo_events
	where 
		dt >= now() - toIntervalDay(6*30) 
		and event in ('view', 'start_task', 'finish_task')
	group by 
		date, 
		user_id
	having started_tasks>0
)
group by date
order by date;
```
###### конверсия в решение задач за час 
```sql
select
	date,
	count() as total_users,
	sum(chain_len_hour=0) as not_solve_tasks,
	sum(chain_len_hour=1) as try_solve_tasks,
	sum(chain_len_hour=2) as success_solve_taks
from (
	select
		toStartOfMonth(dt) as date,
	    user_id,
    	windowFunnel(60*60)(dt, event='start_task', event='finish_task') as chain_len_hour
	from default.demo_events
	where 
		dt >= now() - toIntervalDay(6*30) 
		and event in ('view', 'start_task', 'finish_task')
	group by 
		date, 
		user_id
)
group by date 
order by date
```
### Генератор нагрузки
В состав решения входит генератор нагрузки. Он запускается автоматически и генерерует нагрузку на API в 300 rps. 

Настроить нагрузку можно в [конфиге сервиса](./docker/app.yaml#L10). Значения меньше или равные 0 rps отключают генератор нагрузки.

Насколько API справляется с ней можно оценить в
### Grafana
Логин `admin`, пароль `admin`

web-интерфейс http://127.0.0.1:13000/d/f5kdf5hHz/demo-dashboard?orgId=1&refresh=5s
