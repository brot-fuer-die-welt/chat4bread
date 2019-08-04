package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

func main() {
	var botToken = flag.String("token", "", "Telegram bot token")
	flag.Parse()

	if botToken == nil || *botToken == "" {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		return
	}

	log.Printf("Starting Chat4Bread Backend.")

	// Connect with MongoDB
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	db, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://database:27017"))
	if err != nil {
		log.Panic(err)
	}
	err = db.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Connected to MongoDB database.")

	// Setup state machine
	machine := NewMachine(db)

	// Connect with Telegram
	bot, err := tgbotapi.NewBotAPI(*botToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	// Process messages
	log.Printf("Authorized on Telegram bot account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		reply, err := machine.Generate(update.Message.From.UserName, update.Message.Text)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			reply = fmt.Sprintf("Error: %s", err.Error())
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
	}

	log.Printf("Stopping Chat4Bread Backend.")
}