# go-metr

![Go Report Card](https://goreportcard.com/badge/github.com/vsurkov/go-metr?style=flat-square)

### Запуск БД
docker-compose up -d в директории db
Web-интерфейс будет доступен по адресу http://localhost:8123/play

CREATE TABLE events
(
    sessionid String,
    project String,
    page String,
    loadtime UInt64
) ENGINE = Log;

select * from default.events;
