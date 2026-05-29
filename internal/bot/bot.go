package bot

import (
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/hedgeg0d/agentg/internal/config"
	"github.com/hedgeg0d/agentg/internal/shell"
	"github.com/hedgeg0d/agentg/internal/store"
)

const (
	modeIdle = iota
	modeShell
	modeServiceAdd
)

type Bot struct {
	api   *tgbotapi.BotAPI
	cfg   *config.Config
	store *store.Store
	shell *shell.Manager

	mu       sync.Mutex
	modes    map[int64]int
	monitors map[int64]chan struct{}
}

func New(cfg *config.Config, st *store.Store) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}
	return &Bot{
		api:      api,
		cfg:      cfg,
		store:    st,
		shell:    shell.NewManager(cfg.Timeout()),
		modes:    map[int64]int{},
		monitors: map[int64]chan struct{}{},
	}, nil
}

func (b *Bot) Run() {
	log.Printf("authorized as @%s", b.api.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	for update := range b.api.GetUpdatesChan(u) {
		switch {
		case update.CallbackQuery != nil:
			go b.onCallback(update.CallbackQuery)
		case update.Message != nil:
			go b.onMessage(update.Message)
		}
	}
}

func (b *Bot) authorize(id int64) bool {
	return b.store.ClaimOwner(id)
}

func (b *Bot) mode(chat int64) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.modes[chat]
}

func (b *Bot) setMode(chat int64, m int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.modes[chat] = m
}

func (b *Bot) onMessage(msg *tgbotapi.Message) {
	if !b.authorize(msg.From.ID) {
		b.send(msg.Chat.ID, "⛔ Access denied. This machine already has an owner.")
		return
	}
	if msg.IsCommand() {
		b.onCommand(msg)
		return
	}
	switch b.mode(msg.Chat.ID) {
	case modeShell:
		b.handleShellInput(msg)
	case modeServiceAdd:
		b.handleServiceAdd(msg)
	default:
		b.routeButton(msg)
	}
}

func (b *Bot) onCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		b.setMode(msg.Chat.ID, modeIdle)
		b.sendKeyboard(msg.Chat.ID, "👋 *agentg* ready. Pick an action below.", mainKeyboard())
	case "monitor":
		b.startMonitor(msg.Chat.ID)
	case "services":
		b.showServices(msg.Chat.ID)
	case "shell":
		b.enterShell(msg.Chat.ID)
	case "id":
		b.send(msg.Chat.ID, "Your ID: `"+itoa(msg.From.ID)+"`")
	default:
		b.send(msg.Chat.ID, "Unknown command.")
	}
}

func (b *Bot) routeButton(msg *tgbotapi.Message) {
	switch msg.Text {
	case btnShell:
		b.enterShell(msg.Chat.ID)
	case btnMonitor:
		b.startMonitor(msg.Chat.ID)
	case btnServices:
		b.showServices(msg.Chat.ID)
	case btnStatus:
		b.sendStatus(msg.Chat.ID)
	default:
		b.sendKeyboard(msg.Chat.ID, "Pick an action below.", mainKeyboard())
	}
}

func (b *Bot) onCallback(cb *tgbotapi.CallbackQuery) {
	if !b.authorize(cb.From.ID) {
		b.answer(cb.ID, "Access denied")
		return
	}
	b.routeCallback(cb)
}
