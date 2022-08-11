-- Создание таблицы событий
CREATE TABLE IF NOT EXISTS events (
                                      Timestamp String,
                                      SystemId UUID,
                                      SessionId UUID,
                                      TotalLoading Float64,
                                      DomLoading Float64,
                                      Uri String,
                                      UserAgent String
) engine=Log