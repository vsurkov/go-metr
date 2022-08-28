-- Создание таблицы событий
CREATE DATABASE IF NOT EXISTS rncb;

CREATE TABLE IF NOT EXISTS rncb.events(
                                      Timestamp Int64,
                                      MessageID UUID,
                                      SystemId UUID,
                                      SessionId UUID,
                                      TotalLoading Float64,
                                      DomLoading Float64,
                                      Uri String,
                                      UserAgent String
) engine=Log

-- Удаление таблицы событий
-- DROP TABLE events