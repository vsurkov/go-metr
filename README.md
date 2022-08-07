# go-metr

<<<<<<< HEAD
### Запуск инфраструктуры
docker-compose up -d в директории docker
#### Clickhouse
=======
![Go Report Card](https://goreportcard.com/badge/github.com/vsurkov/go-metr?style=flat-square)

### Запуск БД
docker-compose up -d в директории db
>>>>>>> 866eeb0582df875f309af17a00289196401667ac
Web-интерфейс будет доступен по адресу http://localhost:8123/play

CREATE TABLE events
(
    sessionid String,
    project String,
    page String,
    loadtime UInt64
) ENGINE = Log;

select * from default.events;
<<<<<<< HEAD

#### RabbitMQ
Web-интерфейс будет доступен по адресу http://localhost:15672
креды rabbitmq
=======
>>>>>>> 866eeb0582df875f309af17a00289196401667ac
