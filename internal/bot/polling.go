package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (b *Bot) StartPolling(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	b.log.Info("telegram polling started")

	for {
		select {
		case <-ctx.Done():
			b.log.Info("telegram polling stopped")
			return

		case update, ok := <-updates:
			if !ok {
				b.log.Warn("telegram updates channel closed")
				return
			}

			if err := b.HandleUpdate(ctx, update); err != nil {
				b.log.Error("handle update failed", zap.Error(err))
			}
		}
	}
}
