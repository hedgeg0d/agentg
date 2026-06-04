package bot

import (
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/hedgeg0d/agentg/internal/auth"
	"github.com/hedgeg0d/agentg/internal/config"
	"github.com/hedgeg0d/agentg/internal/shell"
	"github.com/hedgeg0d/agentg/internal/store"
)

const (
	modeIdle = iota
	modeShell
	modeServiceAdd
	modePassword
	modeUserAdd
)

type Bot struct {
	api   *tgbotapi.BotAPI
	cfg   *config.Config
	auth  *auth.Authorizer
	store *store.Store
	shell *shell.Manager

	mu       sync.Mutex
	modes    map[int64]int
	monitors map[int64]chan struct{}
}

func New(cfg *config.Config, az *auth.Authorizer, st *store.Store) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}
	return &Bot{
		api:      api,
		cfg:      cfg,
		auth:     az,
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
	u.AllowedUpdates = []string{"message", "callback_query"}
	for update := range b.api.GetUpdatesChan(u) {
		switch {
		case update.CallbackQuery != nil:
			go b.onCallback(update.CallbackQuery)
		case update.Message != nil:
			go b.onMessage(update.Message)
		}
	}
}

func (b *Bot) principal(from *tgbotapi.User) auth.Principal {
	return auth.Principal{ID: from.ID, Username: from.UserName}
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
	switch b.auth.Check(b.principal(msg.From)) {
	case auth.Denied:
		b.handleDenied(msg)
		return
	case auth.NeedPassword:
		b.handlePassword(msg)
		return
	}
	if b.mode(msg.Chat.ID) == modePassword {
		b.setMode(msg.Chat.ID, modeIdle)
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
	case modeUserAdd:
		b.handleUserAdd(msg)
	default:
		b.routeButton(msg)
	}
}

func (b *Bot) onCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		b.setMode(msg.Chat.ID, modeIdle)
		b.sendKeyboard(msg.Chat.ID, "👋 *agentg* ready. Pick an action below.", mainKeyboard(b.auth.IsAdmin(msg.From.ID)))
	case "monitor":
		b.startMonitor(msg.Chat.ID)
	case "services":
		b.showServices(msg.Chat.ID)
	case "shell":
		b.enterShell(msg.Chat.ID)
	case "fav":
		b.favCommand(msg)
	case "users":
		b.usersCommand(msg)
	case "power":
		b.powerCommand(msg)
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
	case btnUsers:
		b.usersCommand(msg)
	case btnPower:
		b.powerCommand(msg)
	default:
		b.sendKeyboard(msg.Chat.ID, "Pick an action below.", mainKeyboard(b.auth.IsAdmin(msg.From.ID)))
	}
}

func (b *Bot) onCallback(cb *tgbotapi.CallbackQuery) {
	if b.auth.Check(b.principal(cb.From)) != auth.Allowed {
		b.answer(cb.ID, "Access denied")
		return
	}
	b.routeCallback(cb)
}
