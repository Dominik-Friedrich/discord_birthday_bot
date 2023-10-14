package main

import (
	"github.com/spf13/viper"
	"main/src/bot"
	"main/src/features/Player"
)

func main() {
	initViperEnv()

	//db, err := database.NewConnection(
	//	database.Config{
	//		Type:     viper.GetString("database_type"),
	//		Host:     viper.GetString("database_host"),
	//		Port:     viper.GetString("database_port"),
	//		User:     viper.GetString("database_user"),
	//		Password: viper.GetString("database_password"),
	//		Database: viper.GetString("database_database"),
	//	})
	//if err != nil {
	//	log.Panicf("error initializing db connection: %v", err.Error())
	//}

	birthdayBot := bot.NewBot(viper.GetString("discord_token"), viper.GetString("discord_application_id"))

	//birthdayBot.RegisterFeature(Birthday.BirthdayRole(db))
	//birthdayBot.RegisterFeature(Complaint.Complaint(db))
	birthdayBot.RegisterFeature(Player.Player())

	birthdayBot.Run()
}

func initViperEnv() {
	viper.SetEnvPrefix("bot")

	_ = viper.BindEnv("database_type")
	_ = viper.BindEnv("database_host")
	_ = viper.BindEnv("database_port")
	_ = viper.BindEnv("database_user")
	_ = viper.BindEnv("database_password")
	_ = viper.BindEnv("database_database")

	_ = viper.BindEnv("discord_token")
	_ = viper.BindEnv("discord_application_id")

	viper.AutomaticEnv()
}
