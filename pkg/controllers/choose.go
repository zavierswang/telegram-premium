package controllers

import (
	"context"
	"fmt"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"go.uber.org/zap/buffer"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"telegram-premium/pkg/core/cst"
	"telegram-premium/pkg/core/global"
	"telegram-premium/pkg/core/logger"
	"telegram-premium/pkg/middleware"
	"telegram-premium/pkg/models"
	"telegram-premium/pkg/utils"
	"time"
)

func Choose(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	username := callback.From.Username.PeerID()
	userId := callback.From.ID.PeerID()
	chatId := callback.Message.Chat.ID
	messageId := callback.Message.ID
	sess := middleware.SessionManager.Get(ctx)
	sess.Step = middleware.SessionPremiumChoose
	username = strings.ReplaceAll(strings.ToLower(username), "@", "")
	logger.Info("[%s %s] trigger action [choose] callback", userId, username)
	point := utils.RandPoint()
	month, _ := strconv.Atoi(callback.Data)
	TelegramPremium := map[int]float64{
		3:  global.App.Config.Telegram.ThreeMonth,
		6:  global.App.Config.Telegram.SixMonth,
		12: global.App.Config.Telegram.TwelveMonth,
	}
	payments := TelegramPremium[month]
	amount := payments + point
	sess.Amount = amount
	sess.Payments = payments
	sess.Month = month
	logger.Info("[%s %s] choose telegram premium %d month, will pay amount: %.3f USDT", userId, username, sess.Month, sess.Amount)
	return callback.Update.Reply(ctx, callback.Client.EditMessageText(chatId, messageId, "请输入充值的用户名：").ParseMode(tg.HTML))
}

func VoucherChoose(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	messageId := callback.Message.ID
	username := callback.From.Username.PeerID()
	userId := callback.From.ID.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	logger.Info("[%s %s] trigger action [choose] callback", userId, username)
	compile, err := regexp.Compile(`^voucher\s+(?P<balance>\d+)`)
	if err != nil {
		logger.Error("[%s %s] compile balance voucher failed %v", userId, username, err)
		return err
	}
	groups := utils.FindGroups(compile, callback.Data)
	balance, _ := strconv.ParseFloat(groups["balance"], 64)
	balance += utils.RandPoint()
	logger.Info("balance recharge amount of %.3f", balance)
	now := time.Now()

	order := models.Voucher{
		ID:             now.UnixMicro(),
		Balance:        balance,
		Username:       username,
		UserID:         userId,
		MessageID:      messageId,
		CreatedAt:      now,
		Status:         cst.OrderStatusRunning,
		ReceiveAddress: global.App.Config.Telegram.ReceiveAddress,
	}
	tpl := VoucherTmpl{
		ID:             order.ID,
		Balance:        balance,
		Username:       username,
		ReceiveAddress: order.ReceiveAddress,
		ExpiredAt:      now.Add(time.Minute * 10).Format(cst.DateTimeFormatter),
	}
	tmpl, err := template.ParseFiles(cst.VoucherTemplateFile)
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.VoucherTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.VoucherTemplateFile, err)
		return err
	}
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layout.Insert(
		tg.NewInlineKeyboardButtonCallback("取消订单", fmt.Sprintf("cancel voucher %d", now.UnixMicro())),
		tg.NewInlineKeyboardButtonURL("联系客服", fmt.Sprintf("https://t.me/%s", global.App.Config.App.Support)),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
	global.App.DB.Save(&order)
	go tickerTimeout(callback, order, chatId, messageId)
	return callback.Client.EditMessageText(chatId, messageId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
}

type VoucherTmpl struct {
	ID             int64
	Balance        float64
	Username       string
	UserID         string
	ReceiveAddress string
	ExpiredAt      string
}
