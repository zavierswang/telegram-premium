package controllers

import (
	"context"
	"database/sql"
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

func Confirm(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	userId := callback.From.ID.PeerID()
	username := callback.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	logger.Info("[%s %s] trigger action [confirm] callback", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	currentAmount := sess.Amount
	// 检查是否有余额
	var totalBalance sql.NullFloat64
	var totalPayment sql.NullFloat64
	// balance记录总额(balance 帐户余额购买、cash 现金购买)
	global.App.DB.Raw("SELECT SUM(amount) AS amount FROM tb_payment WHERE user_id = ? AND finished = ? AND type = ? ", userId, true, "balance").Scan(&totalPayment)
	// 充值记录总额
	global.App.DB.Raw("SELECT SUM(balance) AS balance FROM tb_voucher WHERE user_id = ? AND status = ?", userId, cst.OrderStatusSuccess).Scan(&totalBalance)
	remainBalance := totalBalance.Float64 - totalPayment.Float64 - currentAmount
	logger.Info("[%s %s] include this order remain balance: %.3f USDT", userId, username, remainBalance)
	return SetPayment(remainBalance, ctx, callback)
}

func Notifier(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	messageId := callback.Message.ID
	userId := callback.From.ID.PeerID()
	username := callback.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	logger.Info("[%s %s] trigger action [notifier] callback", userId, username)
	compile, err := regexp.Compile(`manual\s+(?P<id>\d+)`)
	if err != nil {
		logger.Error("compile manual failed %v", err)
		return err
	}
	var users []models.User
	global.App.DB.First(&users, "user_id = ?", userId)
	if len(users) == 0 || !users[0].IsAdmin {
		logger.Error("illegal user is not administrator, don't click manual button")
		return nil
	}
	groups := utils.FindGroups(compile, callback.Data)
	id := groups["id"]
	var order models.Payment
	err = global.App.DB.First(&order, "id = ?", id).Error
	if err != nil {
		logger.Error("[%s %s] not found user %s", userId, username, id)
		return callback.Client.SendMessage(chatId, "非法用户").DoVoid(ctx)
	}
	tmpl, _ := template.ParseFiles(cst.ManualTemplateFile)
	buf := new(buffer.Buffer)
	tpl := GiftTmpl{
		ForUsername: order.ForUsername,
		Month:       order.Month,
		Amount:      order.Amount,
		Status:      "🟢成功",
		Current:     time.Now().Format(cst.DateTimeFormatter),
	}
	uid, _ := strconv.ParseInt(order.UserID, 10, 64)
	_ = tmpl.Execute(buf, tpl)

	err = callback.Client.SendMessage(tg.ChatID(uid), buf.String()).ParseMode(tg.HTML).DoVoid(ctx)
	if err != nil {
		logger.Error("[%s %s] notifier send message to user failed %v", userId, username, err)
	}
	global.App.DB.Model(models.Payment{}).Where("id = ?", id).Updates(map[string]interface{}{"status": cst.OrderStatusSuccess, "finished": true})
	logger.Info("[%s %s] manual user premium suite successfully", userId, username)
	return callback.Client.EditMessageText(chatId, messageId, "订单已通知到用户~").DoVoid(ctx)
}

func Cancel(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	userId := callback.From.ID.PeerID()
	username := callback.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	logger.Info("[%s %s] trigger action [cancel] callback", userId, username)
	//区分是充值会员订单还是充值余额订单
	compile, err := regexp.Compile(`^cancel\s+(?P<type>\w+)\s+(?P<id>\d+)`)
	if err != nil {
		logger.Error("compile cancel failed %v", err)
		return err
	}
	groups := utils.FindGroups(compile, callback.Data)
	chatId := callback.Message.Chat.ID
	messageId := callback.Message.ID
	sess := middleware.SessionManager.Get(ctx)
	switch groups["type"] {
	case "suite":
		logger.Warn("[%s %s] cancel suite", userId, username)
		var orders []models.Payment
		global.App.DB.Find(&orders, "user_id = ? AND amount = ? AND finished = ? AND expired = ? AND status in ?", userId, sess.Amount, false, false, []int{cst.OrderStatusRunning, cst.OrderStatusReceived, cst.OrderStatusApiSuccess})
		if len(orders) != 0 {
			global.App.DB.Model(models.Payment{}).Delete(orders[0])
		}
	case "voucher":
		logger.Warn("[%s %s] cancel voucher", userId, username)
		var orders []models.Voucher
		global.App.DB.Find(&orders, "id = ? AND status in ?", groups["id"], []int{cst.OrderStatusRunning, cst.OrderStatusReceived, cst.OrderStatusApiSuccess})
		if len(orders) != 0 {
			global.App.DB.Model(models.Voucher{}).Delete(orders[0])
		}
	}
	middleware.SessionManager.Reset(sess)
	logger.Info("[%s %s] telegram premium %s order has been canceled", userId, username, groups["type"])
	_ = callback.Client.DeleteMessage(chatId, messageId).DoVoid(ctx)
	return callback.Client.SendMessage(chatId, "😔很遗憾，您已取消了订单~").DoVoid(ctx)
}

func Close(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	userId := callback.From.ID.PeerID()
	username := callback.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	logger.Info("[%s %s] trigger action [close] callback", userId, username)
	chatId := callback.Message.Chat.ID
	messageId := callback.Message.ID
	sess := middleware.SessionManager.Get(ctx)
	middleware.SessionManager.Reset(sess)
	return callback.Client.DeleteMessage(chatId, messageId).DoVoid(ctx)
}
