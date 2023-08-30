package controllers

import (
	"context"
	"github.com/mr-linch/go-tg/tgb"
	"strings"
	"telegram-premium/pkg/core/logger"
	"telegram-premium/pkg/middleware"
	"telegram-premium/pkg/services"
)

type TMT struct {
	label string
}

func (t *TMT) channel(ctx context.Context, callback *tgb.CallbackQueryUpdate) (string, error) {
	// 走自动充值通道
	userId := callback.From.ID.PeerID()
	username := callback.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	sess := middleware.SessionManager.Get(ctx)
	logger.Info("[%s %s] strategy tmt channel", userId, username)
	err := services.TelegramPremiumSubmit(sess.ForUsername, sess.Payments)
	return t.label, err
}
