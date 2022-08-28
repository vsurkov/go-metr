package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	"strings"
)

func configureParams() {
	viper.AddConfigPath("../../configs/rest-receiver/") //path from config to look config, madness
	viper.AddConfigPath("./config/")                    // path to look for the config file in
	viper.AddConfigPath(".")                            // optionally look for config in the working directory
	viper.SetConfigName("config")                       // Register config file name (no extension)
	viper.SetConfigType("yaml")                         // Look for specific type

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		log.Warn().
			Str("service", helpers.Core).
			Str("method", "configureParams").
			Dict("dict", zerolog.Dict().
				Str("ConfigFileUsed", viper.ConfigFileUsed()).
				Err(err),
			).Msg("can't find config.yml file, starting with default settings \n" +
			"\tconfig should be placed into:\n" +
			"\t'./config/' or '.' path's")

	} else {
		helpers.FailOnError(err, helpers.Core, "config file was found but another error was produced")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("server.port", "3000")
	viper.SetDefault("server.full_name", "rest-application")
	viper.SetDefault("server.name", "rncb")
	viper.SetDefault("server.http_logging", "false")
	viper.SetDefault("server.log_level", 1)
	viper.SetDefault("server.pretty_log", "true")
	viper.SetDefault("server.enable_profiling", "false")
	viper.SetDefault("server.enable_request_id", "false")
	//viper.SetDefault("server.buffer_size", "100")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "9000")
	viper.SetDefault("database.default_database", "default")
	viper.SetDefault("database.user", "default")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.debug", "false")
	viper.SetDefault("database.dial_timeout", "time.Second")
	viper.SetDefault("database.max_open_conns", "10")
	viper.SetDefault("database.max_idle_conns", "5")
	viper.SetDefault("database.conn_max_lifetime", "time.Hour")

	viper.SetDefault("rabbitmq.host", "localhost")
	viper.SetDefault("rabbitmq.port", "5672")
	viper.SetDefault("rabbitmq.user", "rabbitmq")
	viper.SetDefault("rabbitmq.password", "rabbitmq")
	viper.SetDefault("rabbitmq.exchange", "exchange")
	viper.SetDefault("rabbitmq.exchange_type", "direct")
	viper.SetDefault("rabbitmq.binding_key", "key")
	viper.SetDefault("rabbitmq.consumer_tag", "simple-consumer")
}
