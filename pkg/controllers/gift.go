package controllers

import (
	"context"
	"fmt"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"go.uber.org/zap/buffer"
	"html/template"
	"regexp"
	"strings"
	"telegram-premium/pkg/core/cst"
	"telegram-premium/pkg/core/logger"
	"telegram-premium/pkg/middleware"
	"telegram-premium/pkg/services"
	"telegram-premium/pkg/utils"
	"time"
)

type GiftTmpl struct {
	ForUsername string
	Month       int
	Amount      float64
	USDTAddress string
	Mode        string
	Current     string
	Type        string
	Status      string
}

func Gift(ctx context.Context, update *tgb.MessageUpdate) error {
	userId := update.Message.From.ID.PeerID()
	//messageId := update.Message.ID
	username := update.Message.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	logger.Info("[%s %s] trigger action [gift] controller", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	compile, _ := regexp.Compile(`^@?(?P<username>\w+)$`)
	groups := utils.FindGroups(compile, update.Text)
	sess.ForUsername = strings.ToLower(groups["username"])
	forUser := tg.Username(sess.ForUsername)
	link := forUser.Link()
	info, err := services.GetTelegramUserInfo(link)
	if err != nil {
		logger.Error("[%s %s] get telegram userinfo failed %v", userId, username, err)
		return err
	}
	if !info.Exist && info.FirstName == "" {
		logger.Error("[%s %s] not found telegram user %s", userId, username, update.Text)
		return update.Answer("您输入的帐户名错误，请重试~").ParseMode(tg.HTML).DoVoid(ctx)
	}
	now := time.Now()
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layout.Insert(
		tg.NewInlineKeyboardButtonCallback("取消", fmt.Sprintf("cancel suite %d", now.UnixMicro())),
		tg.NewInlineKeyboardButtonCallback("确认支付", "confirm"),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)

	tmpl, err := template.ParseFiles(cst.GiftTemplateFile)
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.GiftTemplateFile, err)
		return err
	}
	tpl := GiftTmpl{
		ForUsername: sess.ForUsername,
		Month:       sess.Month,
		Amount:      sess.Amount,
		Current:     time.Now().Format(cst.DateTimeFormatter),
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if info.Exist && info.LogUrl != "" {
		fileArg := tg.NewFileArgURL(info.LogUrl)
		return update.AnswerPhoto(fileArg).Caption(buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
	}
	return update.Answer(buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
}
