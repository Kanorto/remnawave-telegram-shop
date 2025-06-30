package tribute

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/internal/database"
	"remnawave-tg-shop-bot/internal/remnawave"
	"remnawave-tg-shop-bot/internal/translation"
)

// WebhookEvent represents incoming Tribute webhook payload
type WebhookEvent struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	SentAt    time.Time `json:"sent_at"`
	Payload   struct {
		SubscriptionName string    `json:"subscription_name"`
		SubscriptionID   int64     `json:"subscription_id"`
		PeriodID         int64     `json:"period_id"`
		Period           string    `json:"period"`
		Price            int64     `json:"price"`
		Amount           int64     `json:"amount"`
		Currency         string    `json:"currency"`
		UserID           int64     `json:"user_id"`
		TelegramUserID   int64     `json:"telegram_user_id"`
		ChannelID        int64     `json:"channel_id"`
		ChannelName      string    `json:"channel_name"`
		CancelReason     string    `json:"cancel_reason"`
		ExpiresAt        time.Time `json:"expires_at"`
	} `json:"payload"`
}

// Server handles Tribute webhooks
type Server struct {
	repo         *database.TributeRepository
	customerRepo *database.CustomerRepository
	remClient    *remnawave.Client
	tm           *translation.Manager
	bot          *bot.Bot
}

func NewServer(repo *database.TributeRepository, customerRepo *database.CustomerRepository, remClient *remnawave.Client, tm *translation.Manager, b *bot.Bot) *Server {
	return &Server{repo: repo, customerRepo: customerRepo, remClient: remClient, tm: tm, bot: b}
}

func verifySignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(config.TributeAPIKey()))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if !verifySignature(body, r.Header.Get("trbt-signature")) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	dbEvent := &database.TributeEvent{
		Name:             event.Name,
		CreatedAt:        event.CreatedAt,
		SentAt:           event.SentAt,
		SubscriptionName: event.Payload.SubscriptionName,
		SubscriptionID:   event.Payload.SubscriptionID,
		PeriodID:         event.Payload.PeriodID,
		Period:           event.Payload.Period,
		Price:            event.Payload.Price,
		Amount:           event.Payload.Amount,
		Currency:         event.Payload.Currency,
		UserID:           event.Payload.UserID,
		TelegramUserID:   event.Payload.TelegramUserID,
		ChannelID:        event.Payload.ChannelID,
		ChannelName:      event.Payload.ChannelName,
		ExpiresAt:        event.Payload.ExpiresAt,
	}
	if event.Payload.CancelReason != "" {
		dbEvent.CancelReason = &event.Payload.CancelReason
	}

	if err := s.repo.Insert(r.Context(), dbEvent); err != nil {
		slog.Error("failed to save tribute event", "error", err)
	}

	switch event.Name {
	case "new_subscription":
		s.handleNewSubscription(r.Context(), &event)
	case "cancelled_subscription":
		s.handleCancelledSubscription(r.Context(), &event)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

func (s *Server) handleNewSubscription(ctx context.Context, ev *WebhookEvent) {
	customer, err := s.customerRepo.FindByTelegramId(ctx, ev.Payload.TelegramUserID)
	if err != nil {
		slog.Error("find customer", "err", err)
		return
	}
	if customer == nil {
		customer, err = s.customerRepo.Create(ctx, &database.Customer{TelegramID: ev.Payload.TelegramUserID, Language: "en"})
		if err != nil {
			slog.Error("create customer", "err", err)
			return
		}
	}

	user, err := s.remClient.UpdateExpireAt(ctx, customer.ID, customer.TelegramID, ev.Payload.ExpiresAt)
	if err != nil {
		slog.Error("update remnawave user", "err", err)
		return
	}

	updates := map[string]interface{}{
		"subscription_link": user.SubscriptionUrl,
		"expire_at":         user.ExpireAt,
	}
	if err := s.customerRepo.UpdateFields(ctx, customer.ID, updates); err != nil {
		slog.Error("update customer", "err", err)
	}

	_, err = s.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: customer.TelegramID,
		Text:   s.tm.GetText(customer.Language, "subscription_activated"),
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: s.tm.GetText(customer.Language, "connect_button"), CallbackData: "connect"},
				},
			},
		},
	})
	if err != nil {
		slog.Error("send message", "err", err)
	}
}

func (s *Server) handleCancelledSubscription(ctx context.Context, ev *WebhookEvent) {
	customer, err := s.customerRepo.FindByTelegramId(ctx, ev.Payload.TelegramUserID)
	if err != nil || customer == nil {
		return
	}
	updates := map[string]interface{}{
		"expire_at": ev.Payload.ExpiresAt,
	}
	if err := s.customerRepo.UpdateFields(ctx, customer.ID, updates); err != nil {
		slog.Error("update customer", "err", err)
	}
	s.remClient.UpdateExpireAt(ctx, customer.ID, customer.TelegramID, ev.Payload.ExpiresAt)
}
