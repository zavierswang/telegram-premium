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
	"telegram-premium/pkg/middleware"
	"telegram-premium/pkg/models"
	"time"
)

type Balance struct{}

func (b *Balance) render(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	userId := callback.From.ID.PeerID()
	messageId := callback.Message.ID
	username := callback.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	sess := middleware.SessionManager.Get(ctx)
	logger.Info("[%s %s] premium confirm suite %+v", userId, username, sess)
	logger.Info("[%s %s] strategy balance payment", userId, username)
	order := models.Payment{
		UserID:      userId,
		Username:    username,
		ForUsername: sess.ForUsername,
		Month:       sess.Month,
		Amount:      sess.Amount,
		Payments:    sess.Payments,
		Type:        global.App.Config.Telegram.Channel,
		Mode:        "余额支付",
		MessageID:   messageId,
		Finished:    false,
		Expired:     false,
		Status:      cst.OrderStatusRunning,
		CreatedAt:   time.Now(),
	}
	global.App.DB.Save(&order)
	tpl := GiftTmpl{
		ForUsername: sess.ForUsername,
		Month:       sess.Month,
		Amount:      sess.Amount,
		Mode:        "余额支付",
		USDTAddress: global.App.Config.Telegram.ReceiveAddress,
		Current:     time.Now().Format(cst.DateTimeFormatter),
	}
	tmpl, err := template.ParseFiles(cst.BalanceConfirmTemplateFile)
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.BalanceConfirmTemplateFile, err)
		return err
	}
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](1).Row()
	layout.Insert(tg.NewInlineKeyboardButtonURL("联系客服", fmt.Sprintf("https://t.me/%s", global.App.Config.App.Support)))
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
	buf := new(buffer.Buffer)
	// 对接渠道(manual、tmt)
	label, err := NewChannel(ctx, callback)
	if err != nil {
		logger.Error("[%s %s] balance channel with err: %v", userId, username, err)
		tpl.Status = "🔴失败"
		order.Status = cst.OrderStatusFailure
	} else {
		logger.Info("[%s %s] balance of channel successfully", userId, username)
		tpl.Status = "🟢成功"
		order.Status = cst.OrderStatusSuccess
	}
	middleware.SessionManager.Reset(sess)
	tpl.Type = label
	order.Finished = true
	global.App.DB.Save(&order)
	_ = callback.Client.DeleteMessage(chatId, messageId).DoVoid(ctx)
	_ = tmpl.Execute(buf, tpl)
	if label != "人工充值" {
		_ = callback.Client.SendMessage(chatId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
	} else {
		_ = callback.Client.SendMessage(chatId, "您的订单已收到，请稍等~").ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
	}
	// 群组通知
	for _, group := range global.App.Config.App.Groups {
		gid := tg.Username(group)
		if label == "人工充值" {
			layout = tg.NewButtonLayout[tg.InlineKeyboardButton](1).Row()
			layout.Insert(tg.NewInlineKeyboardButtonCallback("✅已处理并通知用户", fmt.Sprintf("manual %d", order.ID)))
			inlineKeyboard = tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
			_ = callback.Client.SendMessage(gid, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
		} else {
			_ = callback.Client.SendMessage(gid, buf.String()).ParseMode(tg.HTML).DoVoid(ctx)
		}
	}
	return nil
}
