-- Создание таблицы событий
CREATE TABLE IF NOT EXISTS events (
                                      date DateTime,
                                      SystemId UUID,
                                      SessionId UUID,
                                      TotalLoading Float64,
                                      DomLoading Float64,
                                      Uri String,
                                      UserAgent String
) engine=Log