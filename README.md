# go-metr
![Go Report Card](https://goreportcard.com/badge/github.com/vsurkov/go-metr?style=flat-square)

### Запуск инфраструктуры
docker-compose up -d в директории docker
#### Clickhouse
Web-интерфейс будет доступен по адресу http://localhost:8123/play

#### RabbitMQ
Web-интерфейс будет доступен по адресу http://localhost:15672

#### TODO
1. Добавить балансировщик для clickhouse https://www.chproxy.org