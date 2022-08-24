package main

import (
	"github.com/spf13/viper"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	"log"
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
		log.Println("can't find config.yml file, starting with default settings \n" +
			"config should be placed into:\n" +
			"'./config/' or '.' path's")
	} else {
		helpers.FailOnError(err, "config file was found but another error was produced")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("server.port", "3000")
	viper.SetDefault("server.full_name", "rest-application")
	viper.SetDefault("server.name", "rncb")
	viper.SetDefault("server.logging", "false")
	viper.SetDefault("server.enable_profiling", "false")
	viper.SetDefault("server.enable_request_id", "false")
	viper.SetDefault("server.config_path", "./configs/")
	//viper.SetDefault("server.buffer_size", "100")

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

	viper.SetDefault("rabbitmq.host", "localhost")
	viper.SetDefault("rabbitmq.port", "5672")
	viper.SetDefault("rabbitmq.user", "rabbitmq")
	viper.SetDefault("rabbitmq.password", "rabbitmq")
	viper.SetDefault("rabbitmq.exchange", "exchange")
	viper.SetDefault("rabbitmq.exchange_type", "direct")
	viper.SetDefault("rabbitmq.binding_key", "key")
	viper.SetDefault("rabbitmq.consumer_tag", "simple-consumer")
}
