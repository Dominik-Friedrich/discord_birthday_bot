package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"main/src"
	"main/src/commands"
	"main/src/repository"
)

func main() {

	configFile := flag.String("config", "sample.config.json", "Config file to use")
	flag.Parse()

	viper.SetConfigFile(*configFile)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	repo := repository.NewBirthdayRepo()

	birthdayBot := bot.NewBot(viper.GetString("discord.apiToken"))
	birthdayBot.RegisterCommand(commands.AddBirthday(repo))
	//birthdayBot.RegisterCommand(commands.RemoveBirthday())

	birthdayBot.Run()
}
