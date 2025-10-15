package dispatcher

import (
	"context"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/schkedy/go-telegram-dialog/storage"
)

type Dispatcher struct {
	router         Router
	middlewares    []Middleware
	storage        storage.Storage
	count_routines int
}

type Option func(*Dispatcher)

func WithRoutines(count int) Option {
	return func(d *Dispatcher) {
		d.count_routines = count
	}
}

func NewDispatcher(storage storage.Storage, router Router, middlwares []Middleware, opts ...Option) *Dispatcher {
	dp := &Dispatcher{
		storage:        storage,
		router:         router,
		middlewares:    middlwares,
		count_routines: 20, // default value
	}
	for _, opt := range opts {
		opt(dp)
	}

	return dp
}

func (dp *Dispatcher) AddMiddleware(mw ...Middleware) {
	dp.middlewares = append(dp.middlewares, mw...)
}

func (dp *Dispatcher) SetRouter(r Router) {
	dp.router = r
}

func (dp *Dispatcher) StartPolling(ctx context.Context, bot *tgbotapi.BotAPI) error {
	wg := &sync.WaitGroup{}
	dp.storage.Set("key", "value")
	offset := 0
	u := tgbotapi.NewUpdate(offset)
	u.Timeout = 30

	errCh := make(chan error, dp.count_routines)
	updates := bot.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			wg.Add(1)
			go dp.processUpdate(ctx, wg, update, bot, errCh)
		}
		close(errCh)
	}()

	go func() {
		for err := range errCh {
			if err != nil {
				fmt.Println("Error processing update:", err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			bot.StopReceivingUpdates()
			wg.Wait()
			wg.Wait()
			return nil
		default:
			u.Offset = u.Offset + 1
		}
	}

}

func (dp *Dispatcher) processUpdate(ctx context.Context, wg *sync.WaitGroup, update tgbotapi.Update, bot *tgbotapi.BotAPI, errCh chan<- error) {
	defer wg.Done()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	bot.Send(msg)
}
