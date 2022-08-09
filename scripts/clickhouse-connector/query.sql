-- Создание таблицы событий
CREATE TABLE IF NOT EXISTS events (
                                      date DateTime,
                                      systemId UUID,
                                      sessionId UUID,
                                      totalLoading Float64,
                                      domLoading Float64,
                                      uri String,
                                      userAgent String
) engine=Log