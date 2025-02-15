package main

import (
	"log"
	"tarantulago/bot"
	"tarantulago/config"
	"tarantulago/db"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	switch cfg.LogLevel {
	case "debug":
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	case "info":
		log.SetFlags(log.Ldate | log.Ltime)
	}

	database, err := db.NewTarantulaDB(cfg.PostgresURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	tarantulaBot, err := bot.NewTarantulaBot(cfg.TelegramToken, database)
	if err != nil {
		log.Fatal("Failed to create tarantulaBot:", err)
	}

	log.Println("Starting tarantula management tarantulaBot...")
	tarantulaBot.Start()
}
