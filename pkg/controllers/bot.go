package controllers

import "github.com/mr-linch/go-tg"

var Menu = struct {
	Start    string
	Premium  string
	Personal string
	History  string
	Balance  string
}{
	Start:    "🥳 开始",
	Premium:  "💎会员订阅",
	Balance:  "🏦余额充值",
	History:  "🗓️历史订单",
	Personal: "👤个人中心",
}

type Bot struct {
	ReplayMarkup *tg.ReplyKeyboardMarkup
	Cmd          []tg.BotCommand
}

func NewBot() *Bot {
	layout := tg.NewReplyKeyboardMarkup(
		tg.NewButtonRow(
			tg.NewKeyboardButton(Menu.Premium),
			tg.NewKeyboardButton(Menu.Balance),
		),
		tg.NewButtonRow(
			tg.NewKeyboardButton(Menu.History),
			tg.NewKeyboardButton(Menu.Personal),
		),
	)
	layout.ResizeKeyboard = true

	botCmd := []tg.BotCommand{
		{Command: "start", Description: Menu.Start},
	}
	return &Bot{
		ReplayMarkup: layout,
		Cmd:          botCmd,
	}
}
