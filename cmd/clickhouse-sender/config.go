package main

import (
	"github.com/spf13/viper"
	"strings"
)

func configureParams() {
	viper.AddConfigPath("../../configs/clickhouse-sender/") //path from config to look config, madness
	viper.AddConfigPath("./config/")                        // path to look for the config file in
	viper.AddConfigPath(".")                                // optionally look for config in the working directory
	viper.SetConfigName("config")                           // Register config file name (no extension)
	viper.SetConfigType("yaml")                             // Look for specific type

	viper.SetDefault("server.port", "4000")
	viper.SetDefault("server.full_name", "clickhouse-sender")
	viper.SetDefault("server.name", "rncb")
	viper.SetDefault("server.logging", "false")
	viper.SetDefault("server.enable_profiling", "false")
	viper.SetDefault("server.enable_request_id", "false")
	viper.SetDefault("server.config_path", "./configs/")
	viper.SetDefault("server.buffer_size", "1000")

	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", "9000")
	viper.SetDefault("db.default_database", "default")
	viper.SetDefault("db.user", "default")
	viper.SetDefault("db.password", "")
	viper.SetDefault("db.debug", "false")
	viper.SetDefault("db.dial_timeout", "time.Second")
	viper.SetDefault("db.max_open_conns", "10")
	viper.SetDefault("db.max_idle_conns", "5")
	viper.SetDefault("db.conn_max_lifetime", "time.Hour")

	viper.SetDefault("rabbit.host", "localhost")
	viper.SetDefault("rabbit.port", "5672")
	viper.SetDefault("rabbit.user", "rabbitmq")
	viper.SetDefault("rabbit.password", "rabbitmq")
	viper.SetDefault("rabbit.exchange", "exchange")
	viper.SetDefault("rabbit.exchange_type", "direct")
	viper.SetDefault("rabbit.binding_key", "key")
	viper.SetDefault("rabbit.consumer_tag", "simple-consumer")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}
