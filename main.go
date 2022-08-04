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
	"log"
	"time"
)

type Event struct {
	Date      time.Time
	SourceIP  string
	SessionID string
	Project   string
	Page      string
	LoadTime  int64
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
		AppName: "Go-Metr v0.0.0",
	})

	app.Use(logger.New())
	app.Use(requestid.New())

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString(app.Config().AppName)
	})
	app.Get("/status", healthCheck)
	app.Get("/metrics", monitor.New(monitor.Config{Title: "MyService Metrics Page"}))

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

	insertQuery := fmt.Sprintf(`INSERT INTO events (date, source, sessionid, project, page, loadtime) 
		VALUES ('%v','127.0.0.1', '%v', '%v', '%v', %v)`,
		time.Now().Format("20060102150405"),
		body.SessionID,
		body.Project,
		body.Page,
		body.LoadTime)

	err = dbConn.Exec(dbCtx, insertQuery)

	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusOK)
}

// TODO реализовать самодиагностику, посмотри на https://github.com/mackerelio/go-osstat
func healthCheck(ctx *fiber.Ctx) error {
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
			source String,
			sessionid String,
			project String,
			page String,
			loadtime UInt64
		) engine=Log
	`)

	if err != nil {
		return err
	}
	return err
}
