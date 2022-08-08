package main

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"log"
	"time"
)

type Event struct {
	Date         time.Time `json:"date"`
	SystemId     uuid.UUID `json:"systemId"`
	SessionId    uuid.UUID `json:"sessionId"`
	TotalLoading float64   `json:"totalLoading"`
	DomLoading   float64   `json:"domLoading"`
	Uri          string    `json:"uri"`
	UserAgent    string    `json:"userAgent"`
}

var (
	dbConn driver.Conn
	dbCtx  context.Context
)

func main() {

	err := initDB()
	if err != nil {
		log.Fatal(err.Error())
	}

	app := fiber.New(fiber.Config{
		//TODO билдить релиз
		AppName: "go-metr v0.0.0",
	})

	app.Use(logger.New())
	app.Use(requestid.New())

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString(app.Config().AppName)
	})
	app.Get("/status", healthCheck)
	app.Get("/metrics", monitor.New(monitor.Config{Title: "Metrics"}))

	eventApp := app.Group("/event")
	eventApp.Get("", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusForbidden)
	})

	//TODO добавить таймауты и вынести в параметр
	eventApp.Post("", createEvent)

	log.Fatal(app.Listen(":3000"))
}

func createEvent(ctx *fiber.Ctx) error {
	body := new(Event)
	err := ctx.BodyParser(body)

	if err != nil {
		err := ctx.Status(fiber.StatusBadRequest).SendString(err.Error())
		if err != nil {
			return err
		}
		return err
	}

	if err := dbConn.Ping(dbCtx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Printf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// TODO нужны валидаторы!

	selectQuery := fmt.Sprintf("SELECT systemName FROM systems WHERE systemId = '%v'", body.SystemId)
	rows, err := dbConn.Query(dbCtx, selectQuery)

	if err != nil {
		return err
	}

	if !rows.Next() {
		return ctx.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("systemId: %v not found", body.SystemId))
	}

	for rows.Next() {
		var (
			systemName string
		)
		if err := rows.Scan(&systemName); err != nil {
			return err
		}
	}
	err = rows.Close()
	if err != nil {
		return rows.Err()
	}

	return ctx.SendStatus(fiber.StatusOK)
}

// TODO реализовать самодиагностику, посмотри на https://github.com/mackerelio/go-osstat
func healthCheck(ctx *fiber.Ctx) error {

	// Проверка доступности Clickhouse - без базы не сможем принимать метрики (проверка на наличие проекта)
	if err := dbConn.Ping(dbCtx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Printf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return ctx.Status(fiber.StatusInternalServerError).SendString("ERROR")
	}

	return ctx.Status(fiber.StatusOK).SendString("OK")
}

func initDB() error {
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
		return err
	}
	dbConn = conn

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
		return err
	}
	dbCtx = ctx

	if err := conn.Exec(ctx, `DROP TABLE IF EXISTS events`); err != nil {
		return err
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
		return err
	}
	return err
}

//func createTables() error {
//	//CREATE TABLE IF NOT EXISTS systems (
//	//	systemId UUID,
//	//	systemName String
//	//) engine=Log
//}
