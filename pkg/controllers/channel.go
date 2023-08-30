package controllers

import (
	"context"
	"github.com/mr-linch/go-tg/tgb"
	"telegram-premium/pkg/core/global"
)

type AlgoChannel interface {
	channel(ctx context.Context, callback *tgb.CallbackQueryUpdate) (string, error)
}

type Channel struct {
	algo AlgoChannel
}

func (c *Channel) channel(ctx context.Context, callback *tgb.CallbackQueryUpdate) (string, error) {
	return c.algo.channel(ctx, callback)
}

func (c *Channel) set(alg AlgoChannel) {
	c.algo = alg
}

func NewChannel(ctx context.Context, callback *tgb.CallbackQueryUpdate) (string, error) {
	c := &Channel{algo: &TMT{label: "平台充值"}}
	switch global.App.Config.Telegram.Channel {
	case "tmt":
		return c.channel(ctx, callback)
	case "manual":
		c.set(&Manual{label: "人工充值"})
		return c.channel(ctx, callback)
	default:
		return c.channel(ctx, callback)
	}
}
