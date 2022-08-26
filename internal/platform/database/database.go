package database

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vsurkov/go-metr/internal/app/event"
	"github.com/vsurkov/go-metr/internal/common/buffer"
	"github.com/vsurkov/go-metr/internal/common/helpers"
)

type Database struct {
	Config Config
	Conn   driver.Conn
	Ctx    context.Context
	Buffer *buffer.Buffer
}

func (db Database) WriteBatch(mss []event.Event) error {
	batch, err := db.Conn.PrepareBatch(db.Ctx, "INSERT INTO rncb.events (Timestamp, MessageID, SystemId, SessionId, TotalLoading, DomLoading, Uri, UserAgent)")
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

func (db Database) NewConnection(c Config) (*Database, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: c.URI(),
		Auth: clickhouse.Auth{
			Database: c.Database,
			Username: c.User,
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
		log.Error().
			Str("service", helpers.Clickhouse).
			Str("method", "NewConnection").
			Dict("exception", zerolog.Dict().
				Str("URI", c.URI()[0]).
				Str("User", c.User).
				Err(err).
				Stack(),
			).Msg("connecting to database error")
		return &Database{}, err
	}

	err = conn.Close()
	helpers.FailOnError(err, helpers.Clickhouse, "error handled on defer conn.Close() to database")

	ctx := clickhouse.Context(context.Background(), clickhouse.WithSettings(clickhouse.Settings{
		"max_block_size": 10,
	}), clickhouse.WithProgress(func(p *clickhouse.Progress) {
	}), clickhouse.WithProfileInfo(func(p *clickhouse.ProfileInfo) {
	}))
	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Error().
				Str("service", helpers.Clickhouse).
				Str("method", "NewConnection").
				Dict("exception", zerolog.Dict().
					Int32("Code", exception.Code).
					Str("Message", exception.Message).
					Str("StackTrace", exception.StackTrace),
				).Msg("ping fail to database")
		}
		return &Database{}, err
	}

	err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS rncb.events (
			Timestamp Int64,
			MessageID UUID,
			SystemId UUID,
			SessionId UUID,
			TotalLoading Float64,
			DomLoading Float64,
			URI String,			
			UserAgent String
		) engine=Log
	`)

	if err != nil {
		log.Error().
			Str("service", helpers.Clickhouse).
			Str("method", "NewConnection").
			Dict("exception", zerolog.Dict().
				Str("URI", c.URI()[0]).
				Str("User", c.User).
				Err(err).
				Stack(),
			).Msg("error on conn.Exec")
		return &Database{}, err
	}
	return &Database{
		Config: c,
		Conn:   conn,
		Ctx:    ctx,
	}, err
}

func (db Database) GetSystems() (map[string]string, error) {
	var result []struct {
		SystemId   string
		SystemName string
	}
	mp := make(map[string]string)

	if err := db.Conn.Select(db.Ctx, &result, "SELECT * FROM systems"); err != nil {
		log.Error().
			Str("service", helpers.Clickhouse).
			Str("method", "GetSystems").
			Dict("exception", zerolog.Dict().
				Str("URI", db.Config.URI()[0]).
				Str("User", db.Config.User).
				Err(err).
				Stack(),
			).Msg("error while executing database select")
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
			log.Error().
				Str("service", helpers.Clickhouse).
				Str("method", "Ping").
				Dict("exception", zerolog.Dict().
					Int32("Code", exception.Code).
					Str("Message", exception.Message).
					Str("StackTrace", exception.StackTrace),
				).Msg("ping database error")
		}
		return false, err
	}
	return true, nil
}
