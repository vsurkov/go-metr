package main

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/compress"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"log"
	"time"
)

type Database struct {
	conn   driver.Conn
	ctx    context.Context
	buffer *Buffer
}

type dbConfig struct {
	host              string
	port              int
	database          string
	username          string
	password          string
	debug             bool
	dialTimeout       time.Duration
	maxOpenConns      int
	maxIdleConns      int
	connMaxLifetime   time.Duration
	compressionMethod compress.Method
}

var (
	// TODO переписать на поинтеры
	db Database
)

func (db Database) Write(msg *Event) error {
	m := *msg
	insertQuery := fmt.Sprintf(`INSERT INTO events (Timestamp, SystemId, SessionId, TotalLoading, DomLoading, Uri, UserAgent) 
		VALUES ('%v','%v','%v','%v','%v','%v','%v')`,
		m.Timestamp,
		m.SystemId,
		m.SessionId,
		m.TotalLoading,
		m.DomLoading,
		m.Uri,
		m.UserAgent)

	err := db.conn.Exec(db.ctx, insertQuery)

	if err != nil {
		return err
	}
	return nil
}

func (db Database) writeBatch(mss []Event) error {
	batch, err := db.conn.PrepareBatch(db.ctx, "INSERT INTO events (Timestamp, SystemId, SessionId, TotalLoading, DomLoading, Uri, UserAgent)")
	if err != nil {
		return err
	}
	for _, evt := range mss {
		err := batch.Append(
			evt.Timestamp,
			evt.SystemId,
			evt.SessionId,
			evt.TotalLoading,
			evt.DomLoading,
			evt.Uri,
			evt.UserAgent,
		)
		if err != nil {
			return err
		}
	}
	return batch.Send()

}

func (db Database) Connect(c dbConfig) (*Database, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%v:%d", c.host, c.port)},
		Auth: clickhouse.Auth{
			Database: c.database,
			Username: c.username,
			Password: c.password,
		},
		Debug:           c.debug,
		DialTimeout:     c.dialTimeout,
		MaxOpenConns:    c.maxOpenConns,
		MaxIdleConns:    c.maxIdleConns,
		ConnMaxLifetime: c.connMaxLifetime,
		Compression: &clickhouse.Compression{
			Method: c.compressionMethod,
		},
	})
	if err != nil {
		return &Database{}, err
	}

	err = conn.Close()
	failOnError(err, "Error handled on defer conn.Close() to database")

	ctx := clickhouse.Context(context.Background(), clickhouse.WithSettings(clickhouse.Settings{
		"max_block_size": 10,
	}), clickhouse.WithProgress(func(p *clickhouse.Progress) {
	}), clickhouse.WithProfileInfo(func(p *clickhouse.ProfileInfo) {
	}))
	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Printf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return &Database{}, err
	}

	err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS events (
			Timestamp timestamp,
			SystemId UUID,
			SessionId UUID,
			TotalLoading Float64,
			DomLoading Float64,
			Uri String,			
			UserAgent String
		) engine=Log
	`)

	if err != nil {
		return &Database{}, err
	}
	return &Database{
		conn: conn,
		ctx:  ctx,
	}, err
}

func (db Database) Ping() (bool, error) {
	if err := db.conn.Ping(db.ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Printf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return false, err
	}
	return true, nil
}
