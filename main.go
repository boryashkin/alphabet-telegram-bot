package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"math/rand"
	"os"
	"time"
)

const Alphabet = "abcdefghijklmnopqrstuvwxyz"

func main() {
	tgtoken := os.Getenv("TGTOKEN")

	bot, err := tgbotapi.NewBotAPI(tgtoken)
	if err != nil {
		log.Panic(err)
	}

	daemon := Daemon{bot: bot}
	dm := NewDelayMessage()
	rand.Seed(16000000000) //random number
	//bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		update := update // trying to make a local copy to prevent races

		if update.Message == nil {
			log.Println("Not a message")
			continue
		}
		//handle(update.Message)
		//go handle(update.Message)
		daemon.handleWithSingleRsp(&dm, update.Message)
	}
}

type Daemon struct {
	bot *tgbotapi.BotAPI
}

func (d *Daemon) handle(message *tgbotapi.Message) {
	log.Println("handle", message.Text)
	d.send(message.Chat.ID, getNextLetter(message.Text))
}

func (d *Daemon) handleWithSingleRsp(dm *DelayMessage, message *tgbotapi.Message) {
	log.Println("handleWithSingleRsp", message.Text)
	ld := rand.Intn(16000000000)
	dm.SetLastDate(message.Chat.ID, ld)
	go dm.ExecItem(d.send, message.Chat.ID, getNextLetter(message.Text), ld)
}

func (d *Daemon) send(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	_, err := d.bot.Send(msg)
	if err != nil {
		log.Println("Error while sending", err.Error())
	}
}

//
type sendFn func(chatId int64, text string)

func getNextLetter(text string) string {
	reply := Alphabet[0]
	if text != "" {
		for i, letter := range Alphabet {
			if letter == rune(text[0]) {
				reply = Alphabet[(i+1)%len(Alphabet)]
			}
		}
	}

	return string(reply)
}

type DelayMessage struct {
	messages map[int64]int
}

func NewDelayMessage() DelayMessage {
	return DelayMessage{
		messages: make(map[int64]int),
	}
}

func (d *DelayMessage) SetLastDate(chatId int64, random int) {
	log.Println("SetLastDate", chatId, random)
	d.messages[chatId] = random
}

func (d *DelayMessage) ExecItem(fn sendFn, chatId int64, reply string, rand int) {
	time.Sleep(time.Millisecond * 100)
	log.Println("Exec", rand, d.messages[chatId])
	if rand == d.messages[chatId] {
		fn(chatId, reply)
	} else {
		fmt.Println("skip", reply)
	}
}
