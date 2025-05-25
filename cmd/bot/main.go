package main

import (
	"bot"
	"database/sql"
	"db"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not defined")
	}

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Error init Telegram API: %v", err)
	}

	dbConnection, err := sql.Open("sqlite3", "./budgetbot.db")
	if err != nil {
		log.Fatal(err)
	}
	defer dbConnection.Close()

	err = db.InitDB(dbConnection)
	if err != nil {
		log.Fatalf("Error init DB: %v", err)
	}

	database := &db.DB{DB: dbConnection}
	b := bot.NewBot(botAPI, database)

	// setup menu
	b.SetBotCommands()

	log.Printf("Bot started @%s", botAPI.Self.UserName)

	b.HandleUpdates()
}
