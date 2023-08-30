package controllers

import (
	"context"
	"github.com/mr-linch/go-tg/tgb"
	"strings"
	"telegram-premium/pkg/core/logger"
)

type Manual struct {
	label string
}

func (m *Manual) channel(ctx context.Context, callback *tgb.CallbackQueryUpdate) (string, error) {
	// 走人工充值通道
	userId := callback.From.ID.PeerID()
	username := callback.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	logger.Info("[%s %s] strategy manual channel", userId, username)
	return m.label, nil
}
