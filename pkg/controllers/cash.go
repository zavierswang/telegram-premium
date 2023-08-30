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
	"telegram-premium/pkg/services"
	"time"
)

type Cash struct{}

func (c *Cash) render(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	userId := callback.From.ID.PeerID()
	messageId := callback.Message.ID
	username := callback.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	sess := middleware.SessionManager.Get(ctx)
	logger.Info("[%s %s] premium confirm suite %+v", userId, username, sess)
	logger.Info("[%s %s] strategy cash payment", userId, username)
	order := models.Payment{
		UserID:      userId,
		Username:    username,
		ForUsername: sess.ForUsername,
		Month:       sess.Month,
		Amount:      sess.Amount,
		Payments:    sess.Payments,
		Type:        global.App.Config.Telegram.Channel,
		Mode:        "现金支付",
		MessageID:   messageId,
		Finished:    false,
		Expired:     false,
		Status:      cst.OrderStatusRunning,
		CreatedAt:   time.Now(),
	}
	global.App.DB.Save(&order)
	logger.Info("[%s %s] cash payment order id %d", userId, username, order.ID)
	tpl := GiftTmpl{
		ForUsername: sess.ForUsername,
		Month:       sess.Month,
		Amount:      sess.Amount,
		Mode:        "现金支付",
		USDTAddress: global.App.Config.Telegram.ReceiveAddress,
		Current:     time.Now().Format(cst.DateTimeFormatter),
	}
	tmpl, err := template.ParseFiles(cst.CashConfirmTemplateFile)
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.CashConfirmTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.CashConfirmTemplateFile, err)
		return err
	}
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layout.Insert(
		tg.NewInlineKeyboardButtonCallback("取消订单", fmt.Sprintf("cancel suite %d", order.ID)),
		tg.NewInlineKeyboardButtonURL("联系客服", fmt.Sprintf("https://t.me/%s", global.App.Config.App.Support)),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
	go tickerTimeout(callback, order, chatId, messageId)
	link := tg.Username(order.ForUsername)
	info, err := services.GetTelegramUserInfo(link.Link())
	if err != nil {
		logger.Error("[%s %s] %s not found", chatId.PeerID(), username, order.ForUsername)
		return err

	}
	if info.Exist && info.LogUrl != "" {
		return callback.Client.EditMessageCaption(chatId, messageId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
	}
	return callback.Client.EditMessageText(chatId, messageId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
}

func tickerTimeout(callback *tgb.CallbackQueryUpdate, order interface{}, chatId tg.ChatID, messageId int) {
	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			switch order.(type) {
			case models.Voucher:
				record := order.(models.Voucher)
				var orders []models.Voucher
				global.App.DB.Find(&orders, "id = ? AND status in ?", record.ID, []int{2, 3, 4})
				if len(orders) == 0 {
					return
				}
				logger.Warn("[timeout] balance recharge order: %d, has been expired", orders[0].ID)
				//orders[0].Status = "已过期"
				global.App.DB.Delete(orders[0])
				_ = callback.Client.EditMessageText(chatId, messageId, "⏱️您的订单已过期~").ParseMode(tg.HTML).DoVoid(ctx)
			case models.Payment:
				record := order.(models.Payment)
				var orders []models.Payment
				global.App.DB.Find(&orders, "id = ? AND finished = ?", record.ID, false)
				if len(orders) == 0 {
					return
				}
				logger.Warn("[timout] telegram premium suite order: %d, has been expired", orders[0].ID)
				//orders[0].Status = core.OrderStatusCancel
				global.App.DB.Delete(&orders[0])
				_ = callback.Client.DeleteMessage(chatId, messageId).DoVoid(ctx)
				_ = callback.Client.SendMessage(chatId, "⏱️您的订单已过期~").ParseMode(tg.HTML).DoVoid(context.Background())
			}
			return
		}
	}
}
