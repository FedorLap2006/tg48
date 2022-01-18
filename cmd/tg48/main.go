package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"tg48/pkg/mk48"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	dotenv "github.com/joho/godotenv"
)

type Bot struct {
	*tgbotapi.BotAPI
	Mk48 *mk48.API
}

type CommandContext struct {
	*Bot
	Update    tgbotapi.Update
	Message   *tgbotapi.Message
	Arguments []string
}

var (
	leaderboard      map[mk48.LeaderboardPeriod][]mk48.LeaderboardEntry = make(map[mk48.LeaderboardPeriod][]mk48.LeaderboardEntry)
	leaderboardMutex sync.RWMutex
)

type CommandHandlerFunc = func(ctx CommandContext) *tgbotapi.MessageConfig

var leaderboardPeriods = map[string]mk48.LeaderboardPeriod{
	"alltime": mk48.AllTimeLeaderboard,
	"weekly":  mk48.WeeklyTimeLeaderboard,
	"daily":   mk48.DailyTimeLeaderboard,
}

var commands = map[string]CommandHandlerFunc{
	"start": func(ctx CommandContext) *tgbotapi.MessageConfig {
		name := "Theseus"
		if len(ctx.Arguments) > 0 {
			name = ctx.Arguments[0]
		}
		return &tgbotapi.MessageConfig{
			Text: "Welcome home, " + name,
		}
	},
	"leaderboard": func(ctx CommandContext) *tgbotapi.MessageConfig {
		typ := mk48.AllTimeLeaderboard
		if len(ctx.Arguments) > 0 {
			if p, ok := leaderboardPeriods[ctx.Arguments[0]]; ok {
				typ = p
			}
		}
		res := fmt.Sprintf(
			"| %-20s | %-10s |\n|%s|%s|\n",
			"Player", "Score",
			strings.Repeat("-", 22), strings.Repeat("-", 10+2),
		)
		leaderboardMutex.RLock()
		defer leaderboardMutex.RUnlock()
		for _, lb := range leaderboard[typ] {
			padding := utf8.RuneCountInString(lb.Player)
			res += fmt.Sprintf("| %s | %-10d |\n", lb.Player+strings.Repeat(" ", 20-padding), lb.Score)
		}
		return &tgbotapi.MessageConfig{
			Text:      "```\n" + res + "```",
			ParseMode: "MarkdownV2",
		}
	},
}

func handleCommand(bot *Bot, update tgbotapi.Update, msg *tgbotapi.Message) *tgbotapi.MessageConfig {
	arguments := strings.Split(msg.Text, " ")

	if cmd, ok := commands[strings.TrimPrefix(arguments[0], "/")]; ok {
		return cmd(CommandContext{
			Bot:       bot,
			Update:    update,
			Message:   msg,
			Arguments: arguments[1:],
		})
	}

	return nil
}

func handleMessage(bot *Bot, update tgbotapi.Update, msg *tgbotapi.Message) {
	if len(msg.Entities) != 0 && msg.Entities[0].IsCommand() {
		reply := handleCommand(bot, update, msg)
		if reply != nil {
			reply.ReplyToMessageID = msg.MessageID
			reply.ChatID = msg.Chat.ID
			_, err := bot.Send(reply)
			if err != nil {
				log.Panic(err)
			}
		}
	}
}

func handleUpdate(bot *Bot, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handleMessage(bot, update, update.Message)
	}
}

func init() {
	_ = dotenv.Load()
}

func main() {
	tgbot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	client, err := mk48.New()
	if err != nil {
		log.Panic(err)
	}
	go client.Listen()
	defer client.Close()
	bot := &Bot{BotAPI: tgbot, Mk48: client}

	client.Handlers.LeaderboardUpdate = func(lu mk48.LeaderboardUpdate) {
		leaderboardMutex.Lock()
		defer leaderboardMutex.Unlock()
		leaderboard[lu.Period] = lu.Leaderboard
	}

	u := tgbotapi.NewUpdate(0)

	updates := bot.GetUpdatesChan(u)
	log.Println("Bot started!")
	for update := range updates {
		handleUpdate(bot, update)
	}
}
