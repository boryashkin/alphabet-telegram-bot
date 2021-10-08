package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const Alphabet = "abcdefghijklmnopqrstuvwxyz"

const (
	ModeChoosing   = 0
	ModeAlphabet   = 1
	ModeCalculator = 2
)

var ModeWelcomeText = map[int]string{
	ModeChoosing:   "Выберите режим: пришлите 1 для алфавитного, 2 для калькуляторного",
	ModeAlphabet:   "Введите английскую букву, чтобы получить следующую",
	ModeCalculator: "Введите цифры, чтобы получить сумму",
}

func main() {
	tgtoken := os.Getenv("TGTOKEN")

	bot, err := tgbotapi.NewBotAPI(tgtoken)
	if err != nil {
		log.Panic(err)
	}

	dm := NewDelayMessage()
	daemon := Daemon{bot: bot, chatMode: make(map[int64]int, 0), chatCalcSum: make(map[int64]int, 0), dm: &dm}
	rand.Seed(time.Now().UnixNano())
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
		go daemon.handle(update.Message)
	}
}

type Daemon struct {
	bot         *tgbotapi.BotAPI
	chatMode    map[int64]int
	chatCalcSum map[int64]int
	dm          *DelayMessage
}

func (d *Daemon) handle(message *tgbotapi.Message) {
	log.Println("handle", message.Text)
	chosenMode := ModeChoosing
	if mode, ok := d.chatMode[message.Chat.ID]; ok {
		log.Println("ok", message.Text)
		if mode == ModeAlphabet {
			log.Println("ok: ab", message.Text)
			d.send(message.Chat.ID, getNextLetter(message.Text))
			return
		}

		if mode == ModeCalculator {
			log.Println("ok: calc", message.Text)
			d.delayedSend(message.Chat.ID, d.calcNextSum(message.Chat.ID, message.Text))
			return
		}

		if mode == ModeChoosing {
			log.Println("ok: choos", message.Text)
			chosenMode1, err := strconv.Atoi(message.Text)
			chosenMode = chosenMode1 //hack
			if err == nil && (chosenMode == ModeAlphabet || chosenMode == ModeCalculator) {
				log.Println("ok: choos set", message.Text, chosenMode)
				d.send(message.Chat.ID, fmt.Sprintf("Выбран режим %d", chosenMode))
			} else {
				log.Println("ok: else choos", message.Text)
				chosenMode = ModeChoosing
			}
		}
	}

	log.Println("set chosen Mode", message.Text, chosenMode)
	d.chatMode[message.Chat.ID] = chosenMode
	d.send(message.Chat.ID, ModeWelcomeText[chosenMode])
}

func (d *Daemon) delayedSend(chatID int64, text string) {
	log.Println("delayedSend", text)
	ld := rand.Intn(16000000000)
	d.dm.SetLastDate(chatID, ld)
	go d.dm.ExecItem(d.send, chatID, text, ld)
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
func (d *Daemon) calcNextSum(chatID int64, text string) string {
	nextVal, err := strconv.Atoi(text)
	if err != nil {
		return "Введите число"
	}
	d.chatCalcSum[chatID] += nextVal
	reply := fmt.Sprintf("Текущая сумма %d", d.chatCalcSum[chatID])

	return reply
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
