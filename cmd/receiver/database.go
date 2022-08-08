package internal

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"log"
	"time"
)

type Database struct {
	conn driver.Conn
	ctx  context.Context
}

var (
	// TODO переписать на поинтеры
	db Database
)

func (db Database) Connect() (*Database, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
		//Debug:           true,
		DialTimeout:     time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return &Database{}, err
	}

	ctx := clickhouse.Context(context.Background(), clickhouse.WithSettings(clickhouse.Settings{
		"max_block_size": 10,
	}), clickhouse.WithProgress(func(p *clickhouse.Progress) {
		log.Println("progress: ", p)
	}), clickhouse.WithProfileInfo(func(p *clickhouse.ProfileInfo) {
		log.Println("profile info: ", p)
	}))
	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Printf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return &Database{}, err
	}

	if err := conn.Exec(ctx, `DROP TABLE IF EXISTS events`); err != nil {
		return &Database{}, err
	}
	err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS events (
			date DateTime,
			systemId UUID,
			sessionId UUID,
			totalLoading Float64,
			domLoading Float64,
			uri String,			
			userAgent String
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

func (db Database) HasSystem(id uuid.UUID) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("HasSystem verify catch Panic()")
		}
	}()

	ok, err := db.Ping()
	if !ok {
		return false, err
	}

	selectQuery := fmt.Sprintf("SELECT systemName FROM systems WHERE systemId = '%v'", id)
	rows, err := db.conn.Query(db.ctx, selectQuery)

	if err != nil {
		return false, err
	}

	if !rows.Next() {
		return false, nil
	}

	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return false, err
		}
	}

	err = rows.Close()
	if err != nil {
		return false, rows.Err()
	}
	return true, nil
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
