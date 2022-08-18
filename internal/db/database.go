package db

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/compress"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/vsurkov/go-metr/internal/buffer"
	"github.com/vsurkov/go-metr/internal/event"
	"github.com/vsurkov/go-metr/internal/helpers"
	"log"
	"time"
)

type Database struct {
	Conn   driver.Conn
	Ctx    context.Context
	Buffer *buffer.Buffer
}

type DBConfig struct {
	Host              string
	Port              int
	Database          string
	Username          string
	Password          string
	Debug             bool
	DialTimeout       time.Duration
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   time.Duration
	CompressionMethod compress.Method
}

func (db Database) Write(msg *event.Event) error {
	m := *msg
	insertQuery := fmt.Sprintf(`INSERT INTO events (Timestamp, MessageID, SystemId, SessionId, TotalLoading, DomLoading, Uri, UserAgent) 
		VALUES ('%v', '%v','%v','%v','%v','%v','%v','%v')`,
		m.Timestamp,
		m.MessageID,
		m.SystemId,
		m.SessionId,
		m.TotalLoading,
		m.DomLoading,
		m.Uri,
		m.UserAgent)

	err := db.Conn.Exec(db.Ctx, insertQuery)

	if err != nil {
		return err
	}
	return nil
}

func (db Database) WriteBatch(mss []event.Event) error {
	batch, err := db.Conn.PrepareBatch(db.Ctx, "INSERT INTO events (Timestamp, MessageID, SystemId, SessionId, TotalLoading, DomLoading, Uri, UserAgent)")
	if err != nil {
		return err
	}
	for _, evt := range mss {
		err := batch.Append(
			evt.Timestamp,
			evt.MessageID,
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

func (db Database) Connect(c DBConfig) (*Database, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%v:%d", c.Host, c.Port)},
		Auth: clickhouse.Auth{
			Database: c.Database,
			Username: c.Username,
			Password: c.Password,
		},
		Debug:           c.Debug,
		DialTimeout:     c.DialTimeout,
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
		Compression: &clickhouse.Compression{
			Method: c.CompressionMethod,
		},
	})
	if err != nil {
		return &Database{}, err
	}

	err = conn.Close()
	helpers.FailOnError(err, "Error handled on defer conn.Close() to database")

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
			MessageID UUID,
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
		Conn: conn,
		Ctx:  ctx,
	}, err
}

func (db Database) GetSystems() (map[string]string, error) {
	var result []struct {
		SystemId   string
		SystemName string
	}
	mp := make(map[string]string)

	if err := db.Conn.Select(db.Ctx, &result, "SELECT * FROM systems"); err != nil {
		return mp, err
	}

	for i := range result {
		mp[result[i].SystemId] = result[i].SystemName
	}

	return mp, nil
}

func (db Database) Ping() (bool, error) {
	if err := db.Conn.Ping(db.Ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Printf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return false, err
	}
	return true, nil
}
