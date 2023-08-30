package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"go.uber.org/zap/buffer"
	"html/template"
	"strings"
	"telegram-premium/pkg/core/cst"
	"telegram-premium/pkg/core/global"
	"telegram-premium/pkg/core/logger"
	"telegram-premium/pkg/services"
)

func Personal(ctx context.Context, update *tgb.MessageUpdate) error {
	userId := update.Message.From.ID.PeerID()
	firstname := update.From.FirstName
	username := update.Message.From.Username.PeerID()
	username = strings.ReplaceAll(username, "@", "")
	logger.Info("[%s %s] trigger action [personal] controller", userId, username)
	user := update.Message.From.Username
	info, err := services.GetTelegramUserInfo(user.Link())
	if err != nil {
		logger.Error("[%s %s] get telegram userinfo failed %v", userId, username, err)
		return err
	}
	if !info.Exist && info.FirstName == "" {
		logger.Error("[%s %s] not found telegram user", userId, username)
		return update.Answer("您输的帐户名错误，请重试~").ParseMode(tg.HTML).DoVoid(ctx)
	}
	fileArg := tg.NewFileArgURL(info.LogUrl)
	var totalBalance sql.NullFloat64
	var totalPayment sql.NullFloat64
	global.App.DB.Raw("SELECT SUM(amount) AS amount FROM tb_payment WHERE user_id = ? AND expired = ? AND status in ?", userId, false, []int{cst.OrderStatusSuccess}).Scan(&totalPayment)
	global.App.DB.Raw("SELECT SUM(balance) AS balance FROM tb_voucher WHERE user_id = ? AND status = ?", userId, cst.OrderStatusSuccess).Scan(&totalBalance)
	//logger.Info("total balance of user_id %s: %.3f USDT", userId, totalBalance.Float64)
	//logger.Info("total payment of user_id: %s: %.3f USDT", userId, totalPayment.Float64)

	var balance string
	if totalBalance.Float64-totalPayment.Float64 >= 0 {
		balance = fmt.Sprintf("%.3f", totalBalance.Float64-totalPayment.Float64)
	} else {
		balance = "0.00"
	}
	isPremium := "⚫️否"
	if update.From.IsPremium {
		isPremium = "🔵是"
	}
	tpl := PersonalTmpl{
		Username:  strings.ToLower(username),
		Firstname: firstname,
		UserId:    userId,
		IsPremium: isPremium,
		Balance:   balance,
	}
	buf := new(buffer.Buffer)
	tmpl, err := template.ParseFiles(cst.PersonalTemplateFile)
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.PersonalTemplateFile, err)
		return err
	}
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.PersonalTemplateFile, err)
		return err
	}
	if info.Exist && info.LogUrl != "" {
		return update.AnswerPhoto(fileArg).Caption(buf.String()).ParseMode(tg.HTML).DoVoid(ctx)
	}
	return update.Answer(buf.String()).ParseMode(tg.HTML).DoVoid(ctx)
}

type PersonalTmpl struct {
	Username  string
	Firstname string
	UserId    string
	IsPremium string
	Balance   string
}
