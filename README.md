# go-metr

### Запуск инфраструктуры
docker-compose up -d в директории docker
#### Clickhouse
Web-интерфейс будет доступен по адресу http://localhost:8123/play

CREATE TABLE events
(
    sessionid String,
    project String,
    page String,
    loadtime UInt64
) ENGINE = Log;

select * from default.events;

#### RabbitMQ
Web-интерфейс будет доступен по адресу http://localhost:15672
креды rabbitmq