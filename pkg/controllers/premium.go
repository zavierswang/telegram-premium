package controllers

import (
	"context"
	"fmt"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"go.uber.org/zap/buffer"
	"html/template"
	"strings"
	"telegram-premium/pkg/core/cst"
	"telegram-premium/pkg/core/global"
	"telegram-premium/pkg/core/logger"
)

func Premium(ctx context.Context, update *tgb.MessageUpdate) error {
	userId := update.Message.From.ID.PeerID()
	username := update.Message.From.Username.PeerID()
	username = strings.ReplaceAll(username, "@", "")
	logger.Info("[%s %s] trigger action [premium] controller", userId, username)
	buf := new(buffer.Buffer)
	tmpl, err := template.ParseFiles(cst.PremiumTemplateFile)
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.PremiumTemplateFile, err)
		return err
	}
	err = tmpl.Execute(buf, global.App.Config.Telegram.ReceiveAddress)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.PremiumTemplateFile, err)
		return err
	}
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](3).Row()
	layout.Insert(tg.NewInlineKeyboardButtonCallback("3个月", "3"))
	layout.Insert(tg.NewInlineKeyboardButtonCallback("6个月", "6"))
	layout.Insert(tg.NewInlineKeyboardButtonCallback("12个月", "12"))
	layout.Insert(tg.NewInlineKeyboardButtonURL("联系客服", fmt.Sprintf("https://t.me/%s", global.App.Config.App.Support)))
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
	return update.Answer(buf.String()).
		ParseMode(tg.HTML).
		ReplyMarkup(inlineKeyboard).
		DisableWebPagePreview(true).
		DoVoid(ctx)
}
