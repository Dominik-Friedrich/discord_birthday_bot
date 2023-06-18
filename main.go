package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"main/src/bot"
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

	birthdayBot := bot.NewBot(viper.GetString("discord.token"), viper.GetString("discord.application_id"))
	birthdayBot.RegisterFeature(features.BirthdayRole())

	birthdayBot.Run()
}
