package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"

	"remnawave-tg-shop-bot/internal/database"
)

func (h Handler) CreateCustomerIfNotExistMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		var telegramId int64
		var langCode string
		if update.Message != nil {
			telegramId = update.Message.From.ID
			langCode = update.Message.From.LanguageCode
		} else if update.CallbackQuery != nil {
			telegramId = update.CallbackQuery.From.ID
			langCode = update.CallbackQuery.From.LanguageCode
		}

		banned, err := h.bannedRepository.IsBanned(ctx, telegramId)
		if err != nil {
			slog.Error("error checking banned user", err)
			return
		}
		if banned {
			return
		}

		existingCustomer, err := h.customerRepository.FindByTelegramId(ctx, telegramId)
		if err != nil {
			slog.Error("error finding customer by telegram id", err)
			return
		}

		if existingCustomer == nil {
			existingCustomer, err = h.customerRepository.Create(ctx, &database.Customer{
				TelegramID: telegramId,
				Language:   langCode,
			})
			if err != nil {
				slog.Error("error creating customer", err)
				return
			}
		} else {
			updates := map[string]interface{}{
				"language": langCode,
			}

			err = h.customerRepository.UpdateFields(ctx, existingCustomer.ID, updates)
			if err != nil {
				slog.Error("Error updating customer", err)
				return
			}
		}

		if update.Message != nil {
			_ = h.logRepository.Log(context.Background(), telegramId, "message", update.Message.Text)
		} else if update.CallbackQuery != nil {
			_ = h.logRepository.Log(context.Background(), telegramId, "callback", update.CallbackQuery.Data)
		}

		next(ctx, b, update)
	}
}

func (h Handler) BannedMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		var telegramId int64
		if update.Message != nil {
			telegramId = update.Message.From.ID
		} else if update.CallbackQuery != nil {
			telegramId = update.CallbackQuery.From.ID
		}
		banned, err := h.bannedRepository.IsBanned(ctx, telegramId)
		if err != nil {
			slog.Error("error checking banned user", err)
			return
		}
		if banned {
			return
		}
		next(ctx, b, update)
	}
}
