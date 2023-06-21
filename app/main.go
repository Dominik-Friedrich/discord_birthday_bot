package main

import (
	"flag"
	"fmt"
	log "github.com/chris-dot-exe/AwesomeLog"
	"github.com/spf13/viper"
	"main/src/bot"
	"main/src/database"
	"main/src/features"
)

func main() {

	configFile := flag.String("config", "sample.config.json", "Config file to use")
	flag.Parse()

	viper.SetConfigFile(*configFile)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	db, err := database.NewConnection(
		database.Config{
			Type:     viper.GetString("database.type"),
			Host:     viper.GetString("database.host"),
			Port:     viper.GetString("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			Database: viper.GetString("database.database"),
		})
	if err != nil {
		log.Panicf("error initializing db connection: %v", err.Error())
	}

	birthdayBot := bot.NewBot(viper.GetString("discord.token"), viper.GetString("discord.application_id"))

	birthdayBot.RegisterFeature(features.BirthdayRole(db))

	birthdayBot.Run()
}
