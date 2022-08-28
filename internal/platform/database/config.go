package database

import (
	"errors"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/compress"
	"time"
)

type Config struct {
	Host              string
	Port              int
	Database          string
	User              string
	Password          string
	Debug             bool
	DialTimeout       time.Duration
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   time.Duration
	CompressionMethod compress.Method
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.Host == "" {
		return errors.New("database host is required")
	}

	if c.Port == 0 {
		return errors.New("database port is required")
	}

	if c.User == "" {
		return errors.New("database user is required")
	}

	if c.Database == "" {
		return errors.New("database name is required")
	}

	return nil
}

// URI returns a Database driver compatible data source name.
func (c Config) URI() []string {
	addr := make([]string, 0)
	addr = append(addr, fmt.Sprintf("%v:%d", c.Host, c.Port))
	return addr
}
