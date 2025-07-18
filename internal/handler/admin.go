package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"

	"remnawave-tg-shop-bot/internal/config"
)

func (h Handler) StatsCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	total, _ := h.customerRepository.Count(ctx)
	active, _ := h.customerRepository.CountActive(ctx)
	banned, _ := h.bannedRepository.Count(ctx)
	text := fmt.Sprintf(h.translation.GetText("en", "stats_template"), total, banned, active)
	if update.Message != nil {
		lang := update.Message.From.LanguageCode
		if h.translation.HasText(lang, "stats_template") {
			text = fmt.Sprintf(h.translation.GetText(lang, "stats_template"), total, banned, active)
		}
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: text})
		if err != nil {
			slog.Error("send stats", err)
		}
	}
}

func (h Handler) BanCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) < 2 {
		return
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}
	if config.IsAdmin(id) {
		return
	}
	if err = h.bannedRepository.Ban(ctx, id); err != nil {
		slog.Error("ban", err)
	}
	b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(update.Message.From.LanguageCode, "user_banned")})
}

func (h Handler) UnbanCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) < 2 {
		return
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}
	if err = h.bannedRepository.Unban(ctx, id); err != nil {
		slog.Error("unban", err)
	}
	b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(update.Message.From.LanguageCode, "user_unbanned")})
}
