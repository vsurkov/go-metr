package database

//
//import (
//	"context"
//	"fmt"
//	"github.com/ClickHouse/database-go/v2"
//	"github.com/ClickHouse/database-go/v2/lib/compress"
//	"github.com/ClickHouse/database-go/v2/lib/driver"
//	"log"
//	"time"
//)
//
//type Database struct {
//	conn driver.Conn
//	ctx  context.Context
//}
//
//type Config struct {
//	host              string
//	port              int
//	database          string
//	username          string
//	password          string
//	debug             bool
//	dialTimeout       time.Duration
//	maxOpenConns      int
//	maxIdleConns      int
//	connMaxLifetime   time.Duration
//	compressionMethod compress.Method
//}
//
//var (
//	// TODO переписать на поинтеры
//	database Database
//)
//
//func (database Database) NewConnection(c Config) (*Database, error) {
//	conn, err := database.Open(&database.Options{
//		URI: []string{fmt.Sprintf("%v:%d", c.host, c.port)},
//		Auth: database.Auth{
//			Database: c.database,
//			User: c.username,
//			Password: c.password,
//		},
//		Debug:           c.debug,
//		DialTimeout:     c.dialTimeout,
//		MaxOpenConns:    c.maxOpenConns,
//		MaxIdleConns:    c.maxIdleConns,
//		ConnMaxLifetime: c.connMaxLifetime,
//		Compression: &database.Compression{
//			Method: c.compressionMethod,
//		},
//	})
//	if err != nil {
//		return &Database{}, err
//	}
//
//	err = conn.Close()
//	failOnError(err, "Error handled on defer conn.Close() to database")
//
//	ctx := database.Context(context.Background(), database.WithSettings(database.Settings{
//		"max_block_size": 10,
//	}), database.WithProgress(func(p *database.Progress) {
//	}), database.WithProfileInfo(func(p *database.ProfileInfo) {
//	}))
//	if err := conn.Ping(ctx); err != nil {
//		if exception, ok := err.(*database.Exception); ok {
//			log.Printf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
//		}
//		return &Database{}, err
//	}
//
//	//if err := conn.Exec(ctx, `DROP TABLE IF EXISTS events`); err != nil {
//	//	return &Database{}, err
//	//}
//	//err = conn.Exec(ctx, `
//	//	CREATE TABLE IF NOT EXISTS events (
//	//		Date DateTime,
//	//		SystemId UUID,
//	//		SessionId UUID,
//	//		TotalLoading Float64,
//	//		DomLoading Float64,
//	//		URI String,
//	//		UserAgent String
//	//	) engine=Log
//	//`)
//	//
//	//if err != nil {
//	//	return &Database{}, err
//	//}
//	return &Database{
//		conn: conn,
//		ctx:  ctx,
//	}, err
//}
//
//
//
//func (database Database) Ping() (bool, error) {
//	if err := database.conn.Ping(database.ctx); err != nil {
//		if exception, ok := err.(*database.Exception); ok {
//			log.Printf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
//		}
//		return false, err
//	}
//	return true, nil
//}
